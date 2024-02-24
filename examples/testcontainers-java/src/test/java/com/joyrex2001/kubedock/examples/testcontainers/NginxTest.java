package com.joyrex2001.kubedock.examples.testcontainers;

import org.junit.jupiter.api.Test;

import org.testcontainers.containers.BindMode;
import org.testcontainers.containers.GenericContainer;
import org.testcontainers.containers.wait.strategy.Wait;
import org.testcontainers.containers.output.Slf4jLogConsumer;

import org.testcontainers.junit.jupiter.Testcontainers;
import org.testcontainers.utility.DockerImageName;

import static org.assertj.core.api.Assertions.assertThat;
import static org.slf4j.LoggerFactory.getLogger;

import java.net.URI;
import java.net.URL;
import java.io.IOException;

@Testcontainers
public class NginxTest {

    private static final int NGINX_PORT = 8080;
    private static final String NGINX_IMAGE = "nginxinc/nginx-unprivileged"; // "library/nginx"

    @Disabled
    @Test
    @SuppressWarnings("unchecked")
    void testNginx() throws IOException {
        GenericContainer nginx = new GenericContainer(DockerImageName.parse(NGINX_IMAGE))
            //.withFileSystemBind to a folder will copy the folder before the container starts
            .withFileSystemBind("./src/www", "/www", BindMode.READ_ONLY)
            //.withFileSystemBind to a file results into creation of a configmap before the container runs
            .withFileSystemBind("./src/test/resources/nginx.conf", "/etc/nginx/conf.d/default.conf", BindMode.READ_ONLY)
            //.withClasspathResourceMapping results into a copy in a running container (unless kubedock runs with --pre-archive)
            //.withClasspathResourceMapping("nginx.conf", "/etc/nginx/conf.d/default.conf", BindMode.READ_ONLY)
            .withLogConsumer(new Slf4jLogConsumer(getLogger("nginx")))
            .waitingFor(Wait.forHttp("/"))
            .withExposedPorts(NGINX_PORT);

        nginx.start();

        URL serviceUrl = URI.create(String.format("http://%s:%d/", 
                                        nginx.getContainerIpAddress(), 
                                        nginx.getMappedPort(NGINX_PORT))).toURL();

        assertThat(Util.readFromUrl(serviceUrl))
            .contains("<title>Hello!</title>")
            .contains("<h1>Hello!</h1>");

        nginx.stop();
    }
}