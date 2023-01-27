import sys, getopt, os

#convert into executable

def fileExists(path):
    if not os.path.isfile(path):
        print("Please provide a valid file")
        return False
    return True

def main(argv):
    arg = sys.argv[1]
    if arg == "install":
        #install dependencies
        print("installing")
        return 0
    elif arg == "build":
        #complete any compilation
        print("building")
        return 0
    elif arg == "test":
        #run tests
        print("Testing... ")
        os.system("go run ./test/tests.go")
        print("[DONE]")
        return 0
    else:
        #URL file maybe
        if not fileExists(arg): return 1
        print("URL time")
        os.system("go run ./src/go/URLs.go " + arg)
        return 0
    return 1

if __name__ == "__main__":
    main(sys.argv)