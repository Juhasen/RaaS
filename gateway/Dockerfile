FROM nginx:stable-alpine

# Copy custom Nginx configuration for service routing
COPY default.conf /etc/nginx/conf.d/default.conf

CMD ["nginx", "-g", "daemon off;"]
