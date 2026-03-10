# Transpiled from English language source
# Test break statement with repeat forever
print("Testing repeat forever with break")
counter = 0
while True:
    counter = counter + 1
    print(counter)
    if counter == 5:
        print("Breaking out of loop!")
        break
print("Loop exited successfully!")
print("Final counter:")
print(counter)
