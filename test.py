import sys

print open(sys.argv[1]).readline()
with open(sys.argv[1], "w") as f:
    f.write("rand\n")
