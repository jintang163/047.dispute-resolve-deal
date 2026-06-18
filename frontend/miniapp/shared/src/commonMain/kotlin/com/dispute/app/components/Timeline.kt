package com.dispute.app.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.model.MediationProgress

@Composable
fun Timeline(
    modifier: Modifier = Modifier,
    progressList: List<MediationProgress>
) {
    if (progressList.isEmpty()) {
        EmptyCard(
            icon = "📋",
            title = "暂无进度记录",
            description = "调解进度将在这里显示"
        )
        return
    }

    Column(
        modifier = modifier.fillMaxWidth(),
        verticalArrangement = Arrangement.spacedBy(0.dp)
    ) {
        progressList.forEachIndexed { index, progress ->
            TimelineItem(
                progress = progress,
                isLast = index == progressList.lastIndex,
                isActive = index == progressList.lastIndex
            )
        }
    }
}

@Composable
private fun TimelineItem(
    progress: MediationProgress,
    isLast: Boolean,
    isActive: Boolean
) {
    val indicatorColor = if (isActive) {
        MaterialTheme.colorScheme.primary
    } else {
        Color(0xFF22C55E)
    }

    val lineColor = Color(0xFFE5E7EB)
    val textColor = MaterialTheme.colorScheme.onSurface
    val subTextColor = MaterialTheme.colorScheme.onSurfaceVariant

    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.Top
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier
                .width(56.dp)
                .padding(top = 4.dp)
        ) {
            Box(
                modifier = Modifier
                    .size(if (isActive) 20.dp else 16.dp)
                    .background(indicatorColor, CircleShape),
                contentAlignment = Alignment.Center
            ) {
                if (!isActive) {
                    Box(
                        modifier = Modifier
                            .size(6.dp)
                            .background(Color.White, CircleShape)
                    )
                }
            }

            if (!isLast) {
                Box(
                    modifier = Modifier
                        .width(2.dp)
                        .height(56.dp)
                        .background(lineColor)
                        .padding(top = 8.dp)
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 8.dp, bottom = if (isLast) 0.dp else 20.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = progress.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = if (isActive) FontWeight.Bold else FontWeight.SemiBold,
                    color = textColor
                )
                if (isActive) {
                    Box(
                        modifier = Modifier
                            .background(
                                MaterialTheme.colorScheme.primary.copy(alpha = 0.15f),
                                androidx.compose.foundation.shape.RoundedCornerShape(8.dp)
                            )
                            .padding(horizontal = 10.dp, vertical = 4.dp)
                    ) {
                        Text(
                            text = "进行中",
                            color = MaterialTheme.colorScheme.primary,
                            fontSize = 12.sp,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
            }

            if (progress.description != null) {
                Text(
                    text = progress.description!!,
                    style = MaterialTheme.typography.bodyMedium,
                    color = subTextColor,
                    modifier = Modifier.padding(top = 6.dp),
                    lineHeight = 18.sp
                )
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                if (progress.operatorName != null) {
                    Text(
                        text = buildString {
                            append(progress.operatorName)
                            if (progress.operatorRole != null) {
                                append(" · ")
                                append(progress.operatorRole)
                            }
                        },
                        style = MaterialTheme.typography.labelMedium,
                        color = subTextColor
                    )
                }
                Text(
                    text = progress.timestamp,
                    style = MaterialTheme.typography.labelMedium,
                    color = subTextColor
                )
            }
        }
    }
}

@Composable
fun SimpleProgressBar(
    modifier: Modifier = Modifier,
    progress: Float,
    label: String? = null,
    color: Color = MaterialTheme.colorScheme.primary,
    trackColor: Color = Color(0xFFE5E7EB)
) {
    Column(modifier = modifier) {
        if (label != null) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.padding(bottom = 6.dp)
            ) {
                Text(
                    text = label,
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.Medium
                )
                Text(
                    text = "${(progress * 100).toInt()}%",
                    style = MaterialTheme.typography.bodyMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = color
                )
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(8.dp)
                .background(trackColor, androidx.compose.foundation.shape.RoundedCornerShape(4.dp))
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(progress.coerceIn(0f, 1f))
                    .height(8.dp)
                    .background(color, androidx.compose.foundation.shape.RoundedCornerShape(4.dp))
            )
        }
    }
}
