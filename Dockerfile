FROM scratch

COPY l4lb .

ENTRYPOINT [ "/l4lb" ]