package com.joyrex2001.kubedock.examples.testcontainers;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import org.testcontainers.containers.BindMode;
import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.wait.strategy.Wait;
import org.testcontainers.containers.output.Slf4jLogConsumer;

import org.testcontainers.junit.jupiter.Container;
import org.testcontainers.junit.jupiter.Testcontainers;
import org.testcontainers.utility.DockerImageName;

import static org.assertj.core.api.Assertions.assertThat;
import static org.slf4j.LoggerFactory.getLogger;

import java.net.URI;
import java.net.URL;
import java.io.IOException;
import java.net.MalformedURLException;

@Testcontainers
public class NginxTest {

    @Container
    @SuppressWarnings("unchecked")
    public static GenericContainer nginx = new GenericContainer(DockerImageName.parse("library/nginx"))
        //.withFileSystemBind to a folder will copy the folder before the container starts
        .withFileSystemBind("./src/www", "/www", BindMode.READ_ONLY)
        //.withFileSystemBind to a file results into creation of a configmap before the container runs
        .withFileSystemBind("./src/test/resources/nginx.conf", "/etc/nginx/conf.d/default.conf", BindMode.READ_ONLY)
        //.withClasspathResourceMapping results into a copy in a running container
        //.withClasspathResourceMapping("nginx.conf", "/etc/nginx/conf.d/default.conf", BindMode.READ_ONLY)
        .withLogConsumer(new Slf4jLogConsumer(getLogger("nginx")))
        .waitingFor(Wait.forHttp("/"))
        .withExposedPorts(80);

    private static URL serviceUrl;

    @BeforeAll
    static void setUp() throws MalformedURLException {
        nginx.start();
        serviceUrl = URI.create(String.format("http://%s:%d/", 
                                        nginx.getContainerIpAddress(), 
                                        nginx.getMappedPort(80))).toURL();
    }

    @AfterAll
    public static void tearDown() {
        nginx.stop();
    }

    @Test
    void shouldBeStarted() throws IOException {
        assertThat(Util.readFromUrl(serviceUrl))
            .contains("<title>Hello!</title>")
            .contains("<h1>Hello!</h1>");
    }
}