FROM nginx:stable-alpine

# Simple gateway image copying config files into nginx conf.d
COPY config/ /etc/nginx/conf.d/

CMD ["nginx", "-g", "daemon off;"]
