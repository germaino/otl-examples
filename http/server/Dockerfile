FROM ubuntu:18.04


ADD server /usr/local/bin/main
RUN chmod +x /usr/local/bin/main

USER nobody
CMD ["/usr/local/bin/main"]
