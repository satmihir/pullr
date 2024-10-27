# pullr

Pullr is a utility for fast streaming data from S3 pre-signed URLs. Streaming large files from S3 even from EC2 instances in the same region is usually capped at 40mbps. Fortunately, S3 allows range queries to an object and does not limit the number of concurrent pulls. Therefore, a much faster pull is possible if we buffered-ahead parts of the file through separate GET requests. With this, if using an EC2 instance and S3 vpce, the total throughput is only limited by your memory/cpu and network bandwidth of your instance type.

The Go SDK already provides [s3manager](https://pkg.go.dev/github.com/stripe/aws-go/service/s3/s3manager) to accelerate downloads using this technique. However, downloads are a little different from streaming since the order matters and since we are buffering in memory, we can't always just download the whole file ahead of time. Pullr provides an implementation of `io.Reader` backed by a fixed, configurable number of buffer slots that are pulled through separate goroutines and served in order using the `Read([]byte)` method.

The use case is where a large S3 file needs to be streamed for in-memory processing where processing can be much faster than 40mbps pull rate (e.g. in-memory analytics, log processing, log replay simulations etc.). It can also be useful for data transformation jobs where data is streamed in, transformed in some way and uploaded back to S3 through separate goroutines. Being able to pull the stream at 1Gbps+ will likely remove data ingestion from being the bottleneck in such cases.

While it's designed for S3 presigned URLs, it's applicable to any other platform with similar properties.
