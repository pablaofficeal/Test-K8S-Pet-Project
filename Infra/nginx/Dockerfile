FROM nginx:stable-alpine

RUN mkdir -p /var/log/nginx \
    && chown -R nginx:nginx /var/log/nginx

COPY main.conf /etc/nginx/sites-enabled/main.conf
COPY nginx.conf /etc/nginx/nginx.conf