## webauth

A fake webserver with a login page.

### Description

`webauth` runs a minimal webserver, listens for HTTP and HTTPS connections,
and presents a dummy login form. Attempted logins are POST requests. Any
attempt to browse to a different URL is redirected back to the login page,
and it is not possible to actually login (or do anything).

If a GET request is received, the standard login template is executed. If
a POST request (login attempt) is received, the template is rendered with
an extra "failure" string indicating a failed login.
