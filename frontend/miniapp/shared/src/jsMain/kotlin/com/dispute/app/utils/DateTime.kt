package com.dispute.app.utils

import kotlin.js.Date

actual fun formatTimestamp(timestamp: Long): String {
    val date = Date(timestamp.toDouble())
    val month = date.getMonth() + 1
    val day = date.getDate()
    val hours = date.getHours()
    val minutes = date.getMinutes()
    return "%02d-%02d %02d:%02d".format(month, day, hours, minutes)
}
