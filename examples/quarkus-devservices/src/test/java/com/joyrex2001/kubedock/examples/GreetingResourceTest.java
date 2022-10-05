package com.joyrex2001.kubedock.examples;

import io.quarkus.test.junit.QuarkusTest;
import io.quarkus.test.keycloak.client.KeycloakTestClient;
import org.junit.jupiter.api.Test;

import static io.restassured.RestAssured.given;
import static org.hamcrest.CoreMatchers.is;

@QuarkusTest
public class GreetingResourceTest {

    KeycloakTestClient keycloakClient = new KeycloakTestClient();

    @Test
    public void testHelloEndpoint() {
        given()
                .auth().oauth2(getAccessToken("alice"))
                .when().get("/hello")
                .then()
                .statusCode(200)
                .body(is("Hello world!"));
    }

    @Test
    public void testAdminEndpoint() {
        given()
                .auth().oauth2(getAccessToken("bob"))
                .when().get("/admin")
                .then()
                .statusCode(403);
    }

    protected String getAccessToken(String userName) {
        return keycloakClient.getAccessToken(userName);
    }
}