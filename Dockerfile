FROM scratch
MAINTAINER Jonathan Lathrop jonatollah@gmail.com
ADD vaultd vaultd
EXPOSE 8080 8081
ENTRYPOINT ["/vaultd"]