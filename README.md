# dump1090_exporter
Prometheus exporter for dump1090. Tested with [MalcolmRobb/dump1090](https://github.com/MalcolmRobb/dump1090).

## How it works?
dump1090_exporter reads /dump1090/data.json endpoints and exports metrics.

## Example configuration
Example of prometheus.yml
```yaml
scrape_configs:
  - job_name: 'dump1090'
    #scrape_interval: 5s
    static_configs:
      - targets:
        - <host of dump1090>:<port>
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: <host of dump1090_exporter>:<port>
```
