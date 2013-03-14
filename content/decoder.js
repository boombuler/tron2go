function decodeServerMsg(arrBuffer) {
    var buff = function() {
        var idx = 0;
        var buff = new Uint8Array(arrBuffer)

        return {
            isEOF : function() {
                return idx >= buff.length
            },

            read : function() {
                var ret = 0;
                for (var i = 0; ; i++) {
                    var byte = buff[idx++];
                    ret |= (byte & 0x7F) << (7 * i);
                    if ((byte >> 7) == 0)
                        break;
                }
                return ret;
            }
        }
    }();

    var result = '';

    if (buff.isEOF())
        return result;

    var dictionary = [];
    var dictSize = 256;

    for (var i = 0; i < dictSize; i++) {
        dictionary[i] = String.fromCharCode(i);
    }

    var w = String.fromCharCode(buff.read());
    result = w;
    while(!buff.isEOF()) {
        var k = buff.read();
        var entry = '';

        if (dictionary[k]) {
            entry = dictionary[k];
        } else {
            if (k === dictSize) {
                entry = w + w.charAt(0);
            } else {
                return null;
            }
        }

        dictionary[dictSize++] = w + entry.charAt(0);
        w = entry;
        result += entry;
    }
    return decodeURIComponent(escape(result)); // decode UTF8 string
}