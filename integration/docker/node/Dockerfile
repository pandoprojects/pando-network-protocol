FROM ubuntu

COPY ./build/linux/ /bin/

COPY ./integration/docker/node/start.sh /bin/

RUN mkdir /pando

ADD ./integration /pando/integration

VOLUME [ "/data" ]

CMD ["/bin/sh", "-c", "/bin/start.sh"]



