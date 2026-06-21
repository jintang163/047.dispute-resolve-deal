package com.dispute.app.audio

expect fun ByteArray.toBase64(): String

expect fun String.decodeBase64(): ByteArray
