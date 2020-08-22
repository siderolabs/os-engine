# os-engine

## Arch Linux

```bash
sudo pacman -Sy llvm clang linux-headers
git clone https://github.com/iovisor/bcc.git
mkdir bcc/build; cd bcc/build
cmake .. -DCMAKE_INSTALL_PREFIX=/usr
make
sudo make install
```
