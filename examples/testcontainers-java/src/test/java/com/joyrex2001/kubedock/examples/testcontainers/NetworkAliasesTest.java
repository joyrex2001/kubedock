package com.joyrex2001.kubedock.examples.testcontainers;

import org.junit.jupiter.api.Test;
import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.Network;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.io.IOException;

import static org.assertj.core.api.Assertions.assertThat;

@Testcontainers
public class NetworkAliasesTest {

    private static final String ALPINE_IMAGE = "library/alpine";
    private static final int TEST_PORT = 8080;

    @Test
    void testNetworkAliases() throws IOException, InterruptedException {
        Network network = Network.newNetwork();

        GenericContainer foo = new GenericContainer(ALPINE_IMAGE)
                .withNetwork(network)
                .withNetworkAliases("foo")
                // we need to explicitly define the ports we are exposing
                // otherwise the k8s service will not be created.
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