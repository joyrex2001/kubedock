package com.joyrex2001.kubedock.examples;

import io.quarkus.security.Authenticated;

import javax.annotation.security.RolesAllowed;
import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.Produces;
import javax.ws.rs.core.MediaType;

@Authenticated
@Path("/")
public class GreetingResource {

    @GET
    @Produces(MediaType.TEXT_PLAIN)
    @RolesAllowed({"user","admin"})
    @Path("/hello")
    public String say() {
        return "Hello world!";
    }

    @GET
    @Produces(MediaType.TEXT_PLAIN)
    @RolesAllowed({"admin"})
    @Path("/admin")
    public String admin() {
        return "Admin page";
    }

}