import time

def main():
    print("Build worker starting...")
    while True:
        print("Checking for jobs...")
        time.sleep(10)

if __name__ == "__main__":
    main()
