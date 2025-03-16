FROM nginx:1.27.4-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html scripts.js styles.css /usr/share/nginx/html/
COPY favicons/* /usr/share/nginx/html/
HEALTHCHECK --interval=30s --timeout=1s --start-period=3s --retries=3 CMD /usr/bin/curl http://localhost/health/ready | /bin/grep 'Ready check passed' || exit 1