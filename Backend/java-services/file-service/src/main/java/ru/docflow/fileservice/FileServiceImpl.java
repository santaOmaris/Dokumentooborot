package ru.docflow.fileservice;

import com.google.protobuf.ByteString;
import file.FileOuterClass.*;
import file.FileServiceGrpc;
import io.grpc.stub.StreamObserver;
import io.minio.*;
import io.minio.errors.MinioException;

import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.util.UUID;
import java.util.logging.Logger;

public class FileServiceImpl extends FileServiceGrpc.FileServiceImplBase {
    private static final Logger log = Logger.getLogger(FileServiceImpl.class.getName());
    private final MinioClient minio;

    public FileServiceImpl(MinioClient minio) {
        this.minio = minio;
    }

    @Override
    public void uploadFile(UploadFileRequest req, StreamObserver<UploadFileResponse> resp) {
        try {
            ensureBucket(req.getBucket());

            // object_path = UUID + оригинальное расширение (уникально, без коллизий)
            String ext        = extractExt(req.getFilename());
            String objectPath = UUID.randomUUID() + ext;

            byte[] content = req.getContent().toByteArray();
            minio.putObject(PutObjectArgs.builder()
                    .bucket(req.getBucket())
                    .object(objectPath)
                    .stream(new ByteArrayInputStream(content), content.length, -1)
                    .contentType("application/octet-stream")
                    .build());

            resp.onNext(UploadFileResponse.newBuilder().setObjectPath(objectPath).build());
            resp.onCompleted();
        } catch (Exception e) {
            log.severe("uploadFile error: " + e.getMessage());
            resp.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asException());
        }
    }

    @Override
    public void downloadFile(DownloadFileRequest req, StreamObserver<DownloadFileResponse> resp) {
        try {
            InputStream stream = minio.getObject(GetObjectArgs.builder()
                    .bucket(req.getBucket())
                    .object(req.getObjectPath())
                    .build());

            byte[] content = stream.readAllBytes();
            stream.close();

            String filename = extractFilename(req.getObjectPath());
            resp.onNext(DownloadFileResponse.newBuilder()
                    .setContent(ByteString.copyFrom(content))
                    .setFilename(filename)
                    .build());
            resp.onCompleted();
        } catch (MinioException e) {
            log.severe("downloadFile MinIO error: " + e.getMessage());
            resp.onError(io.grpc.Status.NOT_FOUND.withDescription(e.getMessage()).asException());
        } catch (Exception e) {
            log.severe("downloadFile error: " + e.getMessage());
            resp.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asException());
        }
    }

    @Override
    public void deleteFile(DeleteFileRequest req, StreamObserver<DeleteFileResponse> resp) {
        try {
            minio.removeObject(RemoveObjectArgs.builder()
                    .bucket(req.getBucket())
                    .object(req.getObjectPath())
                    .build());

            resp.onNext(DeleteFileResponse.newBuilder().setSuccess(true).build());
            resp.onCompleted();
        } catch (Exception e) {
            log.severe("deleteFile error: " + e.getMessage());
            resp.onError(io.grpc.Status.INTERNAL.withDescription(e.getMessage()).asException());
        }
    }

    // ─── helpers ─────────────────────────────────────────────────────────────

    private void ensureBucket(String bucket) throws Exception {
        boolean exists = minio.bucketExists(BucketExistsArgs.builder().bucket(bucket).build());
        if (!exists) {
            minio.makeBucket(MakeBucketArgs.builder().bucket(bucket).build());
            log.info("created bucket: " + bucket);
        }
    }

    private static String extractExt(String filename) {
        int dot = filename.lastIndexOf('.');
        return (dot >= 0) ? filename.substring(dot) : "";
    }

    private static String extractFilename(String objectPath) {
        int slash = objectPath.lastIndexOf('/');
        return (slash >= 0) ? objectPath.substring(slash + 1) : objectPath;
    }
}
