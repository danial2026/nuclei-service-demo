# Use minimal base image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy pre-built Go binary
COPY nuclei-service-demo .

# Copy Nuclei templates directory into container
COPY templates /templates

# Ensure the non-root user owns the templates directory
RUN chown -R appuser:appuser /templates

# Expose application port
EXPOSE 3742

# Switch to non-root user
USER appuser

# Set environment variable so Nuclei engine knows where templates live
ENV NUCLEI_TEMPLATES_PATH=/templates

# Run the service
CMD ["./nuclei-service-demo"]
