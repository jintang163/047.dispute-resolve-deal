package com.dispute.app.audio

import kotlinx.browser.window

actual fun ByteArray.toBase64(): String {
    val binary = StringBuilder(size)
    for (i in indices) {
        binary.append(this[i].toInt().and(0xFF).toChar())
    }
    return js("btoa(binary.toString())") as String
}

actual fun String.decodeBase64(): ByteArray {
    val binary = js("atob(this)") as String
    val bytes = ByteArray(binary.length)
    for (i in binary.indices) {
        bytes[i] = binary[i].code.toByte()
    }
    return bytes
}
