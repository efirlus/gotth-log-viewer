.PHONY: dev css air templ clean

# Development server with everything running in parallel
dev:
	make -j3 air templ css

# Run air for Go live reload
air:
	air

# Run templ with proxy to our air server
templ:
	templ generate --watch --proxy=http://localhost:5174 --open-browser=false

# Watch and compile CSS with Tailwind
css:
	tailwindcss -i ./internal/view/css/app.css -o ./static/css/styles.css --watch

# Clean up generated files
clean:
	rm -rf tmp/
	rm -f static/css/styles.css
	find . -name "*_templ.go" -type f -delete

css-once:
	tailwindcss -i ./internal/view/css/app.css -o ./static/css/styles.css --minify