File: main
Type: cpu
Time: 2026-02-17 15:32:24 +05
Duration: 60.10s, Total samples = 50ms (0.083%)
Showing nodes accounting for 20ms, 40.00% of 50ms total
flat  flat%   sum%        cum   cum%
-10ms 20.00% 20.00%      -10ms 20.00%  compress/flate.(*huffmanEncoder).bitCounts
10ms 20.00%     0%       10ms 20.00%  runtime.(*limiterEvent).stop
10ms 20.00% 20.00%       10ms 20.00%  runtime.(*mspan).heapBitsSmallForAddr
10ms 20.00% 40.00%       10ms 20.00%  runtime.(*spanSet).push
0     0% 40.00%      -10ms 20.00%  compress/flate.(*Writer).Close (inline)
0     0% 40.00%      -10ms 20.00%  compress/flate.(*compressor).close
0     0% 40.00%      -10ms 20.00%  compress/flate.(*compressor).deflate
0     0% 40.00%       20ms 40.00%  compress/flate.(*compressor).init
0     0% 40.00%      -10ms 20.00%  compress/flate.(*compressor).writeBlock
0     0% 40.00%      -10ms 20.00%  compress/flate.(*huffmanBitWriter).indexTokens
0     0% 40.00%      -10ms 20.00%  compress/flate.(*huffmanBitWriter).writeBlock
0     0% 40.00%      -10ms 20.00%  compress/flate.(*huffmanEncoder).generate
0     0% 40.00%       20ms 40.00%  compress/flate.NewWriter (inline)
0     0% 40.00%      -10ms 20.00%  compress/gzip.(*Writer).Close
0     0% 40.00%       20ms 40.00%  compress/gzip.(*Writer).Write
0     0% 40.00%       20ms 40.00%  encoding/json.(*Encoder).Encode
0     0% 40.00%       10ms 20.00%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0% 40.00%       20ms 40.00%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0% 40.00%       20ms 40.00%  github.com/yandex-practicum/shorten-url/internal/handler.(*Handler).GetURLS
0     0% 40.00%       20ms 40.00%  github.com/yandex-practicum/shorten-url/internal/middleware.(*gzipResponseWriter).Write
0     0% 40.00%       10ms 20.00%  github.com/yandex-practicum/shorten-url/internal/middleware.GzipMiddleware.func1
0     0% 40.00%       20ms 40.00%  main.run.Auth.func3.1
0     0% 40.00%       10ms 20.00%  main.run.RequestLogger.func2.1
0     0% 40.00%       10ms 20.00%  net/http.(*conn).serve
0     0% 40.00%       10ms 20.00%  net/http.HandlerFunc.ServeHTTP
0     0% 40.00%       10ms 20.00%  net/http.serverHandler.ServeHTTP
0     0% 40.00%       10ms 20.00%  runtime.(*mcache).prepareForSweep
0     0% 40.00%       10ms 20.00%  runtime.(*mcache).releaseAll
0     0% 40.00%       10ms 20.00%  runtime.(*mcentral).uncacheSpan
0     0% 40.00%       10ms 20.00%  runtime.(*mspan).typePointersOfUnchecked
0     0% 40.00%       10ms 20.00%  runtime.(*sweepLocked).sweep
0     0% 40.00%       10ms 20.00%  runtime.(*timer).unlockAndRun
0     0% 40.00%       10ms 20.00%  runtime.(*timers).check
0     0% 40.00%       10ms 20.00%  runtime.(*timers).run
0     0% 40.00%       20ms 40.00%  runtime.deductAssistCredit
0     0% 40.00%       20ms 40.00%  runtime.findRunnable
0     0% 40.00%       10ms 20.00%  runtime.forEachP (inline)
0     0% 40.00%       10ms 20.00%  runtime.forEachPInternal
0     0% 40.00%       20ms 40.00%  runtime.gcAssistAlloc
0     0% 40.00%       10ms 20.00%  runtime.gcAssistAlloc.func2
0     0% 40.00%       10ms 20.00%  runtime.gcAssistAlloc1
0     0% 40.00%       10ms 20.00%  runtime.gcDrainN
0     0% 40.00%       10ms 20.00%  runtime.gcMarkDone
0     0% 40.00%       10ms 20.00%  runtime.gcMarkTermination
0     0% 40.00%       10ms 20.00%  runtime.gcMarkTermination.forEachP.func6
0     0% 40.00%       10ms 20.00%  runtime.gcMarkTermination.func4
0     0% 40.00%       10ms 20.00%  runtime.goready (inline)
0     0% 40.00%       10ms 20.00%  runtime.goroutineReady
0     0% 40.00%       10ms 20.00%  runtime.goroutineReady.goready.func1
0     0% 40.00%       20ms 40.00%  runtime.makeslice
0     0% 40.00%       20ms 40.00%  runtime.mallocgc
0     0% 40.00%       10ms 20.00%  runtime.mcall
0     0% 40.00%       10ms 20.00%  runtime.park_m
0     0% 40.00%       10ms 20.00%  runtime.pidleget
0     0% 40.00%       10ms 20.00%  runtime.ready
0     0% 40.00%      -10ms 20.00%  runtime.resetspinning
0     0% 40.00%       10ms 20.00%  runtime.scanobject
0     0% 40.00%       10ms 20.00%  runtime.schedule
0     0% 40.00%       20ms 40.00%  runtime.systemstack