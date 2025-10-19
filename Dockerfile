FROM nginx:alpine

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the 404.json file for NOT_FOUND responses
COPY 404.json /usr/share/nginx/html/404.json

# The output directory will be mounted as a volume at runtime
# This keeps the image small and allows live updates

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
