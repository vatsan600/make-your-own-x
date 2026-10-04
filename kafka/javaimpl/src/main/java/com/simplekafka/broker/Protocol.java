package com.simplekafka.broker;

import java.nio.ByteBuffer;
import java.nio.channels.SocketChannel;

public class Protocol {
    public static final byte PRODUCE = 0x01;
    public static final byte FETCH = 0x02;
    public static final byte METADATA = 0x03;
    public static final byte CREATE_TOPIC = 0x04;

    public static ByteBuffer encodeProduceRequest(String topic, int partition, byte[] message) {
    }

    public static ByteBuffer encodeFetchRequest(String topic, int partition, long offset, int maxBytes) {
    }

    public static ByteBuffer encodeMetadataReqest() {
    }

    public static ByteBuffer encodeCreateTopicRequest(String topic, int numPartitions, short replicationFactor) {
    }

    public static ProduceResult decodeProduceResponse(ByteBuffer buffer) {
    }

    public static FetchResult decodeFetchResponse(bytebuffer buffer) {
    }

    public static MetadataResult decodeMetadataResponse(byteBuffer buffer) {
    }

    public static ByteBuffer encodeReplicateRequest(String topic, int partition, long offset, byte[] message) {
    }

    public static ByteBuffer encodeTopicNotification(String topic) {
    }

    public static void sendErrorResponse(SocketChannel channel, String errorMessage) {}

    public class ProduceResult {
        long offset;
        String error;
    }

    public class FetchResult {
        byte[][] messages;
        String error;
    }

    public class MetadataResult {
        String broker;
        TopicMetadata topic;
        String error;
    }

    public class TopicMetadata {
        string topicName;
        PartitionMetadata partitionMetadata;
    }

    public class PartitionMetadata {
        int partitionId;
        int leaderBrokerId;
        int [] replicaBrokerIds;
    }
}
