import subprocess

result = subprocess.run(["docker", "ps", "-a"], capture_output=True, text=True)

print(result.stdout)
