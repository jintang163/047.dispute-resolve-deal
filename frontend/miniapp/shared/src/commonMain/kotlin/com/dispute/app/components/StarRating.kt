package com.dispute.app.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun StarRating(
    modifier: Modifier = Modifier,
    rating: Int = 0,
    maxStars: Int = 5,
    onRatingChanged: ((Int) -> Unit)? = null,
    starSize: Int = 32,
    readonly: Boolean = false,
    showLabels: Boolean = true
) {
    val filledColor = Color(0xFFFFD700)
    val emptyColor = Color(0xFFE5E7EB)

    val ratingLabels = arrayOf(
        "非常不满意",
        "不满意",
        "一般",
        "满意",
        "非常满意"
    )

    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            repeat(maxStars) { index ->
                val isFilled = index < rating
                val starColor = if (isFilled) filledColor else emptyColor

                Text(
                    text = if (isFilled) "★" else "☆",
                    fontSize = starSize.sp,
                    color = starColor,
                    modifier = Modifier
                        .then(
                            if (!readonly && onRatingChanged != null) {
                                Modifier.clickable { onRatingChanged.invoke(index + 1) }
                            } else {
                                Modifier
                            }
                        )
                        .padding(horizontal = 2.dp)
                )
            }
        }

        if (showLabels && rating > 0) {
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = ratingLabels.getOrNull(rating - 1) ?: "",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.primary,
                fontWeight = FontWeight.SemiBold
            )
        }
    }
}

@Composable
fun RatingBar(
    modifier: Modifier = Modifier,
    initialRating: Int = 0,
    onSubmit: (rating: Int, comment: String) -> Unit,
    onCancel: (() -> Unit)? = null
) {
    var currentRating by remember { mutableStateOf(initialRating) }
    var comment by remember { mutableStateOf(TextFieldValue("")) }

    AppCard(
        modifier = modifier,
        title = "服务评价",
        subtitle = "请对本次调解服务进行评价"
    ) {
        Column {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.Center
            ) {
                StarRating(
                    rating = currentRating,
                    onRatingChanged = { currentRating = it },
                    starSize = 40,
                    showLabels = true
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            val quickTags = listOf(
                "调解员专业",
                "响应速度快",
                "沟通耐心",
                "结果满意",
                "流程便捷",
                "建议改进"
            )

            Text(
                text = "快速选择：",
                style = MaterialTheme.typography.labelLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(bottom = 8.dp)
            )

            var selectedTags by remember { mutableStateOf(setOf<String>()) }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                quickTags.take(3).forEach { tag ->
                    val isSelected = selectedTags.contains(tag)
                    TagCard(
                        text = tag,
                        selected = isSelected,
                        onClick = {
                            selectedTags = if (isSelected) {
                                selectedTags - tag
                            } else {
                                selectedTags + tag
                            }
                            comment = TextFieldValue(
                                (selectedTags + tag).joinToString("、")
                            )
                        }
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                quickTags.drop(3).forEach { tag ->
                    val isSelected = selectedTags.contains(tag)
                    TagCard(
                        text = tag,
                        selected = isSelected,
                        onClick = {
                            selectedTags = if (isSelected) {
                                selectedTags - tag
                            } else {
                                selectedTags + tag
                            }
                            comment = TextFieldValue(
                                (selectedTags + tag).joinToString("、")
                            )
                        }
                    )
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            Text(
                text = "详细评价：",
                style = MaterialTheme.typography.labelLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(bottom = 8.dp)
            )

            OutlinedTextField(
                value = comment,
                onValueChange = { comment = it },
                placeholder = {
                    Text(
                        "请输入您的详细评价，帮助我们改进服务...",
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp),
                shape = RoundedCornerShape(12.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = MaterialTheme.colorScheme.primary,
                    unfocusedBorderColor = Color(0xFFE5E7EB)
                ),
                maxLines = 4
            )

            Spacer(modifier = Modifier.height(24.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                if (onCancel != null) {
                    androidx.compose.material3.Button(
                        onClick = onCancel,
                        modifier = Modifier
                            .weight(1f)
                            .height(48.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = androidx.compose.material3.ButtonDefaults.buttonColors(
                            containerColor = MaterialTheme.colorScheme.surfaceVariant,
                            contentColor = MaterialTheme.colorScheme.onSurface
                        )
                    ) {
                        Text(
                            text = "取消",
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }

                androidx.compose.material3.Button(
                    onClick = { onSubmit(currentRating, comment.text) },
                    modifier = Modifier
                        .weight(2f)
                        .height(48.dp),
                    shape = RoundedCornerShape(12.dp),
                    enabled = currentRating > 0
                ) {
                    Text(
                        text = "提交评价",
                        fontWeight = FontWeight.SemiBold,
                        fontSize = 16.sp
                    )
                }
            }
        }
    }
}
