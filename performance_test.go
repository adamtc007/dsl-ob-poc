package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	registry "dsl-ob-poc/internal/domain-registry"
	hedgefundinvestor "dsl-ob-poc/internal/domains/hedge-fund-investor"
	"dsl-ob-poc/internal/domains/onboarding"
)

// PerformanceMetrics holds performance measurement data
type PerformanceMetrics struct {
	TotalRequests     int
	SuccessfulReqs    int
	FailedReqs        int
	TotalDuration     time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	AvgLatency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	RequestsPerSecond float64
	Latencies         []time.Duration
}

// TestMultiDomainConcurrency tests concurrent access to multi-domain system
func TestMultiDomainConcurrency(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("ConcurrentRouting", func(t *testing.T) {
		concurrency := 100
		requestsPerWorker := 50
		totalRequests := concurrency * requestsPerWorker

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalRequests)
		errors := make(chan error, totalRequests)

		messages := []string{
			"Create investor opportunity for concurrent test",
			"Start KYC process for investor",
			"Create case for CBU-CONCURRENT-TEST",
			"Add custody and fund accounting products",
			"Begin compliance verification",
			"Set banking instructions",
			"Submit subscription request",
		}

		start := time.Now()

		// Launch concurrent workers
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < requestsPerWorker; j++ {
					reqStart := time.Now()

					req := &registry.RoutingRequest{
						Message:   messages[j%len(messages)],
						SessionID: fmt.Sprintf("worker-%d-req-%d", workerID, j),
						Context:   make(map[string]interface{}),
						Timestamp: time.Now(),
					}

					_, err := router.Route(ctx, req)
					reqDuration := time.Since(reqStart)

					if err != nil {
						errors <- err
					} else {
						results <- reqDuration
					}
				}
			}(i)
		}

		// Wait for all workers to complete
		wg.Wait()
		totalDuration := time.Since(start)

		// Collect results
		close(results)
		close(errors)

		var latencies []time.Duration
		for latency := range results {
			latencies = append(latencies, latency)
		}

		var errorCount int
		for range errors {
			errorCount++
		}

		metrics := calculateMetrics(latencies, totalDuration)

		t.Logf("Concurrent Routing Performance:")
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Successful: %d", len(latencies))
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Total Duration: %v", totalDuration)
		t.Logf("  Requests/Second: %.2f", metrics.RequestsPerSecond)
		t.Logf("  Average Latency: %v", metrics.AvgLatency)
		t.Logf("  Min Latency: %v", metrics.MinLatency)
		t.Logf("  Max Latency: %v", metrics.MaxLatency)
		t.Logf("  P95 Latency: %v", metrics.P95Latency)
		t.Logf("  P99 Latency: %v", metrics.P99Latency)

		// Performance assertions
		if metrics.RequestsPerSecond < 1000 {
			t.Logf("Warning: Low throughput: %.2f requests/second", metrics.RequestsPerSecond)
		}

		if metrics.AvgLatency > 10*time.Millisecond {
			t.Logf("Warning: High average latency: %v", metrics.AvgLatency)
		}

		if errorCount > 0 {
			t.Errorf("Errors occurred during concurrent testing: %d", errorCount)
		}
	})

	t.Run("ConcurrentDomainAccess", func(t *testing.T) {
		concurrency := 50
		requestsPerWorker := 100
		totalRequests := concurrency * requestsPerWorker

		var wg sync.WaitGroup
		results := make(chan time.Duration, totalRequests)
		errors := make(chan error, totalRequests)

		domains := []string{"hedge-fund-investor", "onboarding"}

		start := time.Now()

		// Launch concurrent domain access workers
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < requestsPerWorker; j++ {
					reqStart := time.Now()

					domainName := domains[j%len(domains)]
					_, err := reg.Get(domainName)

					reqDuration := time.Since(reqStart)

					if err != nil {
						errors <- err
					} else {
						results <- reqDuration
					}
				}
			}(i)
		}

		wg.Wait()
		totalDuration := time.Since(start)

		// Collect results
		close(results)
		close(errors)

		var latencies []time.Duration
		for latency := range results {
			latencies = append(latencies, latency)
		}

		var errorCount int
		for range errors {
			errorCount++
		}

		metrics := calculateMetrics(latencies, totalDuration)

		t.Logf("Concurrent Domain Access Performance:")
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Successful: %d", len(latencies))
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Requests/Second: %.2f", metrics.RequestsPerSecond)
		t.Logf("  Average Latency: %v", metrics.AvgLatency)

		if errorCount > 0 {
			t.Errorf("Errors occurred during concurrent domain access: %d", errorCount)
		}
	})
}

// TestMemoryPerformance tests memory usage and garbage collection impact
func TestMemoryPerformance(t *testing.T) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("MemoryAllocationPattern", func(t *testing.T) {
		iterations := 10000

		// Pre-allocate to reduce GC impact during measurement
		requests := make([]*registry.RoutingRequest, iterations)
		for i := 0; i < iterations; i++ {
			requests[i] = &registry.RoutingRequest{
				Message:   fmt.Sprintf("Memory test request %d", i),
				SessionID: fmt.Sprintf("mem-test-%d", i),
				Context:   make(map[string]interface{}),
				Timestamp: time.Now(),
			}
		}

		// Measure memory allocation during routing
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_, err := router.Route(ctx, requests[i])
			if err != nil {
				t.Errorf("Routing failed at iteration %d: %v", i, err)
			}
		}

		duration := time.Since(start)

		t.Logf("Memory Performance Test:")
		t.Logf("  Iterations: %d", iterations)
		t.Logf("  Total Duration: %v", duration)
		t.Logf("  Average per request: %v", duration/time.Duration(iterations))
		t.Logf("  Requests per second: %.2f", float64(iterations)/duration.Seconds())
	})
}

// TestScalabilityLimits tests system behavior under extreme loads
func TestScalabilityLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability test in short mode")
	}

	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("HighVolumeRouting", func(t *testing.T) {
		volumes := []int{1000, 5000, 10000, 25000}

		for _, volume := range volumes {
			t.Run(fmt.Sprintf("Volume_%d", volume), func(t *testing.T) {
				var latencies []time.Duration
				start := time.Now()

				for i := 0; i < volume; i++ {
					reqStart := time.Now()

					req := &registry.RoutingRequest{
						Message:   fmt.Sprintf("High volume test request %d", i),
						SessionID: fmt.Sprintf("hv-test-%d", i),
						Context:   make(map[string]interface{}),
						Timestamp: time.Now(),
					}

					_, err := router.Route(ctx, req)
					if err != nil {
						t.Errorf("Routing failed at iteration %d: %v", i, err)
						continue
					}

					latencies = append(latencies, time.Since(reqStart))
				}

				totalDuration := time.Since(start)
				metrics := calculateMetrics(latencies, totalDuration)

				t.Logf("Volume %d Results:", volume)
				t.Logf("  Requests/Second: %.2f", metrics.RequestsPerSecond)
				t.Logf("  Average Latency: %v", metrics.AvgLatency)
				t.Logf("  P95 Latency: %v", metrics.P95Latency)
				t.Logf("  P99 Latency: %v", metrics.P99Latency)

				// Check for performance degradation
				if metrics.P99Latency > 50*time.Millisecond {
					t.Logf("Warning: P99 latency is high at volume %d: %v", volume, metrics.P99Latency)
				}
			})
		}
	})
}

// TestLongRunningStability tests system stability over extended periods
func TestLongRunningStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long running test in short mode")
	}

	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	t.Run("ExtendedOperation", func(t *testing.T) {
		duration := 30 * time.Second // Run for 30 seconds
		requestInterval := 10 * time.Millisecond

		start := time.Now()
		var requestCount int
		var errorCount int
		var latencies []time.Duration

		ticker := time.NewTicker(requestInterval)
		defer ticker.Stop()

		timeout := time.After(duration)

		for {
			select {
			case <-timeout:
				// Test duration completed
				goto TestComplete
			case <-ticker.C:
				reqStart := time.Now()
				requestCount++

				req := &registry.RoutingRequest{
					Message:   fmt.Sprintf("Long running test request %d", requestCount),
					SessionID: fmt.Sprintf("lr-test-%d", requestCount),
					Context:   make(map[string]interface{}),
					Timestamp: time.Now(),
				}

				_, err := router.Route(ctx, req)
				reqLatency := time.Since(reqStart)

				if err != nil {
					errorCount++
					t.Logf("Request %d failed: %v", requestCount, err)
				} else {
					latencies = append(latencies, reqLatency)
				}

				// Check memory usage periodically
				if requestCount%1000 == 0 {
					metrics := reg.GetMetrics()
					if !reg.IsHealthy() {
						t.Errorf("Registry became unhealthy at request %d", requestCount)
						return
					}
					t.Logf("Health check at request %d: %d domains, healthy=%t",
						requestCount, metrics.TotalDomains, reg.IsHealthy())
				}
			}
		}

	TestComplete:
		totalDuration := time.Since(start)
		perfMetrics := calculateMetrics(latencies, totalDuration)

		t.Logf("Long Running Stability Test Results:")
		t.Logf("  Test Duration: %v", duration)
		t.Logf("  Total Requests: %d", requestCount)
		t.Logf("  Successful: %d", len(latencies))
		t.Logf("  Failed: %d", errorCount)
		t.Logf("  Error Rate: %.2f%%", float64(errorCount)/float64(requestCount)*100)
		t.Logf("  Average RPS: %.2f", perfMetrics.RequestsPerSecond)
		t.Logf("  Average Latency: %v", perfMetrics.AvgLatency)
		t.Logf("  P99 Latency: %v", perfMetrics.P99Latency)

		// Stability assertions
		if errorCount > requestCount/100 { // More than 1% error rate
			t.Errorf("High error rate during stability test: %.2f%%", float64(errorCount)/float64(requestCount)*100)
		}

		if !reg.IsHealthy() {
			t.Error("Registry is not healthy after long running test")
		}
	})
}

// calculateMetrics computes performance metrics from latency measurements
func calculateMetrics(latencies []time.Duration, totalDuration time.Duration) *PerformanceMetrics {
	if len(latencies) == 0 {
		return &PerformanceMetrics{}
	}

	// Sort latencies for percentile calculations
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)

	// Simple bubble sort for time.Duration
	for i := 0; i < len(sortedLatencies); i++ {
		for j := 0; j < len(sortedLatencies)-1-i; j++ {
			if sortedLatencies[j] > sortedLatencies[j+1] {
				sortedLatencies[j], sortedLatencies[j+1] = sortedLatencies[j+1], sortedLatencies[j]
			}
		}
	}

	// Calculate basic metrics
	var total time.Duration
	min := sortedLatencies[0]
	max := sortedLatencies[len(sortedLatencies)-1]

	for _, latency := range latencies {
		total += latency
	}

	avg := total / time.Duration(len(latencies))

	// Calculate percentiles
	p95Index := int(float64(len(sortedLatencies)) * 0.95)
	if p95Index >= len(sortedLatencies) {
		p95Index = len(sortedLatencies) - 1
	}

	p99Index := int(float64(len(sortedLatencies)) * 0.99)
	if p99Index >= len(sortedLatencies) {
		p99Index = len(sortedLatencies) - 1
	}

	rps := float64(len(latencies)) / totalDuration.Seconds()

	return &PerformanceMetrics{
		TotalRequests:     len(latencies),
		SuccessfulReqs:    len(latencies),
		TotalDuration:     totalDuration,
		MinLatency:        min,
		MaxLatency:        max,
		AvgLatency:        avg,
		P95Latency:        sortedLatencies[p95Index],
		P99Latency:        sortedLatencies[p99Index],
		RequestsPerSecond: rps,
		Latencies:         latencies,
	}
}

// BenchmarkDomainOperations provides benchmark tests for key operations
func BenchmarkDomainOperations(b *testing.B) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	b.Run("RouterRoute", func(b *testing.B) {
		req := &registry.RoutingRequest{
			Message:   "Create investor opportunity benchmark",
			SessionID: "benchmark-session",
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := router.Route(ctx, req)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RegistryGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			domainName := "hedge-fund-investor"
			if i%2 == 0 {
				domainName = "onboarding"
			}
			_, err := reg.Get(domainName)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GetAllVocabularies", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = reg.GetAllVocabularies()
		}
	})

	b.Run("DomainValidateVerbs", func(b *testing.B) {
		domain, _ := reg.Get("hedge-fund-investor")
		dsl := "(investor.start-opportunity :legal-name \"Test\" :type \"INDIVIDUAL\")"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = domain.ValidateVerbs(dsl)
		}
	})

	b.Run("VocabularyAccess", func(b *testing.B) {
		domain, _ := reg.Get("onboarding")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = domain.GetVocabulary()
		}
	})
}

// BenchmarkConcurrentOperations tests concurrent performance
func BenchmarkConcurrentOperations(b *testing.B) {
	reg := registry.NewRegistry()
	reg.Register(hedgefundinvestor.NewDomain())
	reg.Register(onboarding.NewDomain())
	router := registry.NewRouter(reg)
	ctx := context.Background()

	b.Run("ConcurrentRouting", func(b *testing.B) {
		req := &registry.RoutingRequest{
			Message:   "Concurrent benchmark routing test",
			SessionID: "concurrent-benchmark",
			Context:   make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := router.Route(ctx, req)
				if err != nil {
					b.Error(err)
				}
			}
		})
	})

	b.Run("ConcurrentDomainAccess", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				domainName := "hedge-fund-investor"
				if i%2 == 0 {
					domainName = "onboarding"
				}
				_, err := reg.Get(domainName)
				if err != nil {
					b.Error(err)
				}
				i++
			}
		})
	})
}
