import sjcl from 'sjcl'

sjcl.codec.arrayBuffer = { // Codec not included by default in sjcl library
  toBits: function (buffer) {
    var i, out=[], len, inView, tmp;

    if (buffer.byteLength === 0) {
      return [];
    }

    inView = new DataView(buffer);
    len = inView.byteLength - inView.byteLength%4;

    for (var i = 0; i < len; i+=4) {
      out.push(inView.getUint32(i));
    }

    if (inView.byteLength%4 != 0) {
      tmp = new DataView(new ArrayBuffer(4));
      for (var i = 0, l = inView.byteLength%4; i < l; i++) {
        //we want the data to the right, because partial slices off the starting bits
        tmp.setUint8(i+4-l, inView.getUint8(len+i)); // big-endian, 
      }
      out.push(
        sjcl.bitArray.partial( (inView.byteLength%4)*8, tmp.getUint32(0) )
      ); 
    }

    return out;
  },
};

export function sha256Hash(arrayBuffer) {
  const bitArray = sjcl.codec.arrayBuffer.toBits(arrayBuffer);
  const out = sjcl.hash.sha256.hash(bitArray);
  const hash = sjcl.codec.hex.fromBits(out);
  
  return hash;
}