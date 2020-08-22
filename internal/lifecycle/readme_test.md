This code block with one command should be picked up:
[//]: # ({sst-run-unix})
```bash
echo hello world
```

This code block with two commands with indents should be picked up:
[//]: # ({sst-run-unix})
```bash
echo line one
    echo line two
```

This code block without a comment code tag should not be picked up:
```bash
echo hello world
```
