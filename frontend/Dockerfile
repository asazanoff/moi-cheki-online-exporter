FROM nginx:1.27.3-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html scripts.js styles.css /usr/share/nginx/html/
COPY qr.html /usr/share/nginx/html/