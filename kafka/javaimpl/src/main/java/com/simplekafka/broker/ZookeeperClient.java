package com.simplekafka.broker;

public class ZookeeperClient {
    public void connect() throws IOException, InterruptedException {
        zookeeper = new Zookeeper(getConnectString(), SESSION_TIMEOUT, this);
        connectedSignal.await();

        // Create required paths if they don't exist
        createPath("/brokers");
        createPath("/topics");
        createPath("/controllers");
    }
    
}
