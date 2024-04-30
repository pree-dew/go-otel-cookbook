### go-otel-cookbooks

This repo describes recipes for pushing instrumented code metrics via OTel to a backend.
The aim of this repo is to showcase the flexibility of OTel metric ingestion pipeline
and educate the user on various approaches so that they can take an informed decision on
which approach to follow for their specific use case.

### Push using direct remotewrite - OTel backend

- Metric ingestion flow

  <insert diagram here>

- Use when you want to:

  - Push directly to an OTel supported backend. Here backend means the storage system that you are going to query e.g. Prometheus.

- Do not use when:

  - Your backend does not support native OTel based ingestion.

- Pros:

  - No intermediate agents e.g. vmagent, prometheus-agent are required.

- Cons:

  - If you already have a vmagent running to push metrics to a backend, as this approach eliminates
    the need for an intermediatry, it also means that you have to learn OTel's metric flow better
    and account for OTel's metrics when debugging ingestion issues. You have to replace your vmagent
    debugging metrics by OTel metrics.

- Try it out

  ```
  cd push-using-direct-remotewrite/direct-backend

  # Read main.go for understanding the instrumentation done

  # To run a sample setup of the flow
  docker-compose up
  ```

### Push using direct remotewrite - via collectior

- Metric ingestion flow

  <insert diagram here>
