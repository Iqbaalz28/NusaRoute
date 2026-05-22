import os
import re
import glob

def process_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    original_content = content

    # Add imports
    imports_to_add = [
        '"sync"',
        '"time"',
        '"github.com/prometheus/client_golang/prometheus/promhttp"'
    ]
    for imp in imports_to_add:
        if imp not in content:
            content = re.sub(r'import \(', f'import (\n\t{imp}', content, count=1)

    # Inject Prometheus Route
    if 'promhttp.Handler()' not in content:
        content = re.sub(r'(mux := http\.NewServeMux\(\))', r'\1\n\tmux.Handle("/metrics", promhttp.Handler())', content)

    # Inject Metrics Middleware
    if 'middleware.Metrics(' not in content:
        # find the last middleware application or the assignment h = mux
        if 'h = middleware.Recovery(h)' in content:
            content = content.replace('h = middleware.Recovery(h)', 'h = middleware.Recovery(h)\n\th = middleware.Metrics(h)')
        elif 'var h http.Handler = mux' in content:
            content = content.replace('var h http.Handler = mux', 'var h http.Handler = mux\n\th = middleware.Metrics(h)')

    # Add WaitGroup initialization
    if 'var wg sync.WaitGroup' not in content:
        if 'defer cancel()' in content:
            content = content.replace('defer cancel()', 'defer cancel()\n\tvar wg sync.WaitGroup')
        else:
            # fallback if no context exists
            content = re.sub(r'(mux := http\.NewServeMux\(\))', r'ctx, cancel := context.WithCancel(context.Background())\n\tdefer cancel()\n\tvar wg sync.WaitGroup\n\t\1', content)

    # Replace Graceful shutdown block
    # We look for a pattern that matches the end of the server initialization
    shutdown_replacement = """	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info(context.Background(), fmt.Sprint("Shutting down service gracefully..."))
		cancel() // notify workers to stop

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Info(context.Background(), fmt.Sprintf("Server shutdown error: %v", err))
		}
	}()

	logger.Info(context.Background(), fmt.Sprintf("✅ Service running on port %s", port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(fmt.Sprintf("Server failed: %v", err))
	}

	wg.Wait()
	logger.Info(context.Background(), fmt.Sprint("All workers stopped. Goodbye!"))
"""
    
    # Remove old graceful shutdown and ListenAndServe
    # We use regex to match from "// Graceful shutdown" up to the end of ListenAndServe block
    pattern = r'// Graceful shutdown.*?if err := http\.ListenAndServe.*?\n\t\}'
    if re.search(pattern, content, re.DOTALL):
        content = re.sub(pattern, shutdown_replacement, content, flags=re.DOTALL)
    elif 'http.ListenAndServe' in content:
        # If no Graceful shutdown comment exists, just replace ListenAndServe
        pattern2 = r'logger\.Info\(.*?running on port.*?\n\s*if err := http\.ListenAndServe.*?\n\t\}'
        content = re.sub(pattern2, shutdown_replacement, content, flags=re.DOTALL)

    if content != original_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Updated {filepath}")

# Find all main.go
files = glob.glob('back-end/services/*/cmd/main.go') + glob.glob('back-end/api-gateway/cmd/main.go')
for f in files:
    process_file(f)

