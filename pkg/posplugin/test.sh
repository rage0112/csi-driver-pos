sudo ./csc identity plugin-info --endpoint "unix://tmp/csi.sock"
sudo ./csc controller new moon --endpoint "unix:///tmp/csi.sock" --cap 1,block, --req-bytes 2147483648 --params server=127.0.0.1,share=/
sudo ./csc controller list-volumes --endpoint "unix:///tmp/csi.sock"
