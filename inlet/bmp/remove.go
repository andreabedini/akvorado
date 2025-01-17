// SPDX-FileCopyrightText: 2022 Free Mobile
// SPDX-License-Identifier: AGPL-3.0-only

package bmp

import (
	"context"
	"time"
)

func (c *Component) peerRemovalWorker() error {
	for {
		select {
		case <-c.t.Dying():
			return nil
		case pkey := <-c.peerRemovalChan:
			exporterStr := pkey.exporter.Addr().Unmap().String()
			for {
				// Do one run of removal (read/write lock)
				removed, done, duplicate := func() (int, bool, bool) {
					start := c.d.Clock.Now()
					ctx, cancel := context.WithTimeout(c.t.Context(context.Background()),
						c.config.RIBPeerRemovalMaxTime)
					c.mu.Lock()
					defer func() {
						cancel()
						c.mu.DowngradeLock()
						c.metrics.locked.WithLabelValues("peer-removal").Observe(
							float64(c.d.Clock.Now().Sub(start).Nanoseconds()) / 1000 / 1000 / 1000)
					}()
					pinfo := c.peers[pkey]
					if pinfo == nil {
						// Already removed (removal can be queued several times)
						return 0, true, true
					}
					removed, done := c.rib.flushPeerContext(ctx, pinfo.reference,
						c.config.RIBPeerRemovalBatchRoutes)
					if done {
						// Run was complete, remove the peer (we need the lock)
						delete(c.peers, pkey)
					}
					return removed, done, false
				}()

				// Update stats and optionally sleep (read lock)
				func() {
					defer c.mu.RUnlock()
					c.metrics.routes.WithLabelValues(exporterStr).Sub(float64(removed))
					if done {
						// Run was complete, update metrics
						if !duplicate {
							c.metrics.peers.WithLabelValues(exporterStr).Dec()
							c.metrics.peerRemovalDone.WithLabelValues(exporterStr).Inc()
						}
						return
					}
					// Run is incomplete, update metrics and sleep a bit
					c.metrics.peerRemovalPartial.WithLabelValues(exporterStr).Inc()
					select {
					case <-c.t.Dying():
						done = true
					case <-time.After(c.config.RIBPeerRemovalSleepInterval):
					}
				}()
				if done {
					break
				}
			}
		}
	}
}
