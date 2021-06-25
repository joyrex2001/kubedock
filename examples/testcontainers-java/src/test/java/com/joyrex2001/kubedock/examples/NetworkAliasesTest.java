package com.joyrex2001.kubedock.examples.testcontainers;

import org.junit.jupiter.api.Test;

import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.Network;

import org.testcontainers.junit.jupiter.Testcontainers;

import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.lang.InterruptedException;

@Testcontainers
public class NetworkAliasesTest {

    private static final String ALPINE_IMAGE = "library/alpine"; // "library/nginx"
    private static final int TEST_PORT = 8080;

    @Test
    void shouldBeStarted() throws IOException, InterruptedException {
        Network network = Network.newNetwork();

        GenericContainer foo = new GenericContainer(ALPINE_IMAGE)
                .withNetwork(network)
                .withNetworkAliases("foo")
                .withExposedPorts(TEST_PORT)
                .withCommand("/bin/sh", "-c", "while true ; do printf 'HTTP/1.1 200 OK\\n\\nyay' | nc -l -p 8080; done");

        GenericContainer bar = new GenericContainer(ALPINE_IMAGE)
                .withNetwork(network)
                .withCommand("top");

        foo.start();
        bar.start();
    
        String response = bar.execInContainer("wget", "-O", "-", "http://foo:8080").getStdout();
        assertThat(response).contains("yay");

        foo.stop();
        bar.stop();
    }
}