FROM prom/prometheus
COPY production.yml /etc/prometheus/prometheus.yml
CMD ["--config.file=/etc/prometheus/prometheus.yml", "--enable-feature=expand-external-labels", "--enable-feature=agent"]