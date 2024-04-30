### go-otel-cookbooks

This repo describes recipes for pushing instrumented code metrics via OTel to a backend.
The aim of this repo is to showcase the flexibility of OTel metric ingestion pipeline
and educate the user on various approaches so that they can take an informed decision on
which approach to follow for their specific use case.

### Pre-requisite environment variables set

    ```
    export ENDPOINT_DOMAIN='http://otel-agent:4317'
    export REMOTE_WRITE_URL='http://victoriametrics:8429/api/v1/write'
    ```

### Push using direct remotewrite - OTel backend

- Metric ingestion flow

  - insert diagram here

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

  docker-compose up
  ```

  This setup currently pushes to a vmagent based backend, because the test setup doesn't have an OTel backend to write to.

### Push using direct remotewrite - via collectior

- Metric ingestion flow

  - insert diagram here

- Use when you want to:

  - Push to a backend that supports integration with an agent (e.g. vmagent), but may not necessarily
    be speaking the OTel protocol.

- Do not use when:
  - Your setup does not have an existing agent the OTel collector can write to.

- Pros:
  - Leverage your existing vmagent setup to remove the scrape functionality but still keep the remote write functionality.
  - Transparent to the remote write endpoint. This can be used as an intermediate step to move away from running extra agents.

- Cons:
  - Requires maintaining both OTel and vmagent setups.

- Try it out

  ```
  cd push-using-direct-remotewrite/via-collector

  docker-compose up
  ```

### Push using agent - vmagent

- Metric ingestion flow

  - insert diagram here

- How is this different from pushing via direct remote write collector scenario described above?
  - When using 


