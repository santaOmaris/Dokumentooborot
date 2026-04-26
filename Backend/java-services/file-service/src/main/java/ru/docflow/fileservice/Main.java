package ru.docflow.fileservice;

import io.grpc.Server;
import io.grpc.netty.shaded.io.grpc.netty.NettyServerBuilder;
import io.minio.MinioClient;

import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

public class Main {
    private static final Logger log = Logger.getLogger(Main.class.getName());

    public static void main(String[] args) throws Exception {
        String endpoint  = getEnv("MINIO_ENDPOINT",   "http://localhost:9000");
        String accessKey = getEnv("MINIO_ACCESS_KEY", "minioadmin");
        String secretKey = getEnv("MINIO_SECRET_KEY", "minioadmin");
        int    grpcPort  = Integer.parseInt(getEnv("GRPC_PORT", "9091"));

        MinioClient minioClient = MinioClient.builder()
                .endpoint(endpoint)
                .credentials(accessKey, secretKey)
                .build();

        Server server = NettyServerBuilder
                .forPort(grpcPort)
                .addService(new FileServiceImpl(minioClient))
                .build()
                .start();

        log.info("file-service gRPC started on port " + grpcPort);

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            log.info("file-service: shutting down");
            server.shutdown();
            try {
                server.awaitTermination(10, TimeUnit.SECONDS);
            } catch (InterruptedException ignored) {}
        }));

        server.awaitTermination();
    }

    private static String getEnv(String key, String fallback) {
        String v = System.getenv(key);
        return (v != null && !v.isBlank()) ? v : fallback;
    }
}
