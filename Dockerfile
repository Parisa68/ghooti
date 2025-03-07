FROM alpine

COPY input.txt /input.txt

CMD ["sh", "-c", "cat /input.txt"]
