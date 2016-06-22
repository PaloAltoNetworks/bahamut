# bahamut
Bahamut is the materia for building a server

# Dealing with Client Certificates

Bahamut implements a mutual TLS authentication. This means that clients must provide valid certificates.
Within the context of the Aporeto implementation, in order to download a client certificate in your 
browser, you need the following steps. 

1. Get the client certificates created by the infrastructure scripts (cert.p12 ) 

2. Get the CA certificate (private) used by the infrastructure scripts

3. In MacOS drop the CA certificate in your Keychain. Double-click on the certificate, select "Trust" and "When using this Certificate: Always Trust"

4. Drop the client certificate in your Keychain. This will require a password.

5. When you try to access Squall (or any other server that uses mutual TLS) from your browser, a window will pop-up and ask you to select the certificate that you want to use. 
