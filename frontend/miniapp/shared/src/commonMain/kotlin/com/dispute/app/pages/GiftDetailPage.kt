package com.dispute.app.pages

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dispute.app.LocalAppState
import com.dispute.app.LocalApiClient
import com.dispute.app.LocalRouter
import com.dispute.app.Route
import com.dispute.app.components.AppCard
import com.dispute.app.model.Gift
import com.dispute.app.model.GiftExchangeRecord
import kotlinx.coroutines.launch

@Composable
fun GiftDetailPage() = GiftDetailContent()

@Composable
private fun GiftDetailContent() {
    val appState = LocalAppState.current
    val router = LocalRouter.current
    val apiClient = LocalApiClient.current

    val gift = appState.selectedGift
    val gridWorker = appState.gridWorker
    var exchangeQuantity by remember { mutableStateOf(1) }
    var isExchanging by remember { mutableStateOf(false) }
    var showExchangeSuccess by remember { mutableStateOf(false) }

    val canExchange = gift != null && gridWorker != null &&
            gridWorker.points >= gift.points * exchangeQuantity &&
            (gift.stock == null || gift.stock >= exchangeQuantity) &&
            !isExchanging

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.background)
    ) {
        TopBar(
            title = "礼品详情",
            onBack = { router.back() }
        )

        if (gift == null) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Text("礼品信息不存在", style = MaterialTheme.typography.bodyLarge)
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .verticalScroll(rememberScrollState())
            ) {
                GiftImageSection(gift = gift)

                GiftBasicInfoSection(gift = gift)

                ExchangeRulesSection()

                GiftDescriptionSection(gift = gift)
            }

            BottomExchangeSection(
                gift = gift,
                userPoints = gridWorker?.points ?: 0,
                exchangeQuantity = exchangeQuantity,
                onQuantityChange = { exchangeQuantity = it },
                canExchange = canExchange,
                isExchanging = isExchanging,
                onExchangeClick = {
                    isExchanging = true
                    appState.appScope.launch {
                        try {
                            val request = com.dispute.app.api.ExchangeGiftRequest(
                                workerId = "gw001",
                                giftId = gift.id,
                                quantity = exchangeQuantity
                            )
                            val response = apiClient.gridWorker.exchangeGift(request)
                            appState.addExchangeRecord(response)

                            val updatedWorker = gridWorker?.copy(
                                points = gridWorker.points - gift.points * exchangeQuantity
                            )
                            if (updatedWorker != null) {
                                appState.setGridWorker(updatedWorker)
                            }

                            showExchangeSuccess = true
                        } catch (e: Exception) {
                            appState.showToast("兑换失败: ${e.message}")
                        } finally {
                            isExchanging = false
                        }
                    }
                }
            )
        }
    }

    if (showExchangeSuccess) {
        ExchangeSuccessDialog(
            gift = gift,
            quantity = exchangeQuantity,
            onDismiss = {
                showExchangeSuccess = false
                router.back()
            }
        )
    }
}

@Composable
private fun TopBar(title: String, onBack: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(56.dp)
            .background(Color.White)
            .padding(horizontal = 16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "←",
            fontSize = 24.sp,
            modifier = Modifier.clickable(onClick = onBack)
        )
        Spacer(modifier = Modifier.width(8.dp))
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold
        )
    }
}

@Composable
private fun GiftImageSection(gift: Gift) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(240.dp)
            .background(
                Brush.linearGradient(
                    colors = getGradientColors(gift.categoryId)
                )
            ),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(gift.icon, fontSize = 80.sp)
            Spacer(modifier = Modifier.height(12.dp))
            Box(
                modifier = Modifier
                    .background(Color.White.copy(alpha = 0.9f), RoundedCornerShape(16.dp))
                    .padding(horizontal = 16.dp, vertical = 8.dp)
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("⭐", fontSize = 18.sp)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "${gift.points} 积分",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFFF9800)
                    )
                }
            }
        }
    }
}

@Composable
private fun GiftBasicInfoSection(gift: Gift) {
    AppCard(
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = gift.name,
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )

            Text(
                text = gift.description,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(4.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                if (gift.stock != null) {
                    InfoItem(
                        icon = "📦",
                        label = "库存",
                        value = if (gift.stock > 0) "${gift.stock}件" else "已售罄"
                    )
                }

                InfoItem(
                    icon = "🏷️",
                    label = "分类",
                    value = gift.categoryName
                )

                gift.expiryDays?.let { days ->
                    InfoItem(
                        icon = "📅",
                        label = "有效期",
                        value = "${days}天"
                    )
                }
            }
        }
    }
}

@Composable
private fun InfoItem(
    icon: String,
    label: String,
    value: String
) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Text(icon, fontSize = 16.sp)
        Spacer(modifier = Modifier.width(4.dp))
        Text(
            text = "$label: ",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium
        )
    }
}

@Composable
private fun ExchangeRulesSection() {
    AppCard(
        title = "兑换说明",
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 6.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            RuleItem(text = "兑换后礼品将在3个工作日内发放")
            RuleItem(text = "积分不足可通过完成任务获取更多积分")
            RuleItem(text = "已兑换礼品不可退换，请确认后兑换")
            RuleItem(text = "如有问题请联系客服: 400-123-4567")
        }
    }
}

@Composable
private fun RuleItem(text: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        verticalAlignment = Alignment.Top
    ) {
        Text(
            text = "•",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(top = 1.dp)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = text,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun GiftDescriptionSection(gift: Gift) {
    AppCard(
        title = "礼品详情",
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = gift.longDescription,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurface
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "使用方法:",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold
            )

            Text(
                text = gift.usageMethod,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun BottomExchangeSection(
    gift: Gift,
    userPoints: Int,
    exchangeQuantity: Int,
    onQuantityChange: (Int) -> Unit,
    canExchange: Boolean,
    isExchanging: Boolean,
    onExchangeClick: () -> Unit
) {
    val totalPoints = gift.points * exchangeQuantity
    val pointsDiff = totalPoints - userPoints

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(Color.White)
            .padding(horizontal = 16.dp, vertical = 16.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("我的积分:", style = MaterialTheme.typography.bodyMedium)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = userPoints.toString(),
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        color = if (userPoints >= totalPoints) Color(0xFF22C55E) else Color(0xFFEF4444)
                    )
                    if (pointsDiff > 0) {
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = "(还差$pointsDiff)",
                            style = MaterialTheme.typography.bodySmall,
                            color = Color(0xFFEF4444)
                        )
                    }
                }

                QuantitySelector(
                    quantity = exchangeQuantity,
                    minQuantity = 1,
                    maxQuantity = gift.stock ?: 99,
                    onQuantityChange = onQuantityChange
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Column {
                    Text(
                        text = "兑换积分",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("⭐", fontSize = 16.sp)
                        Spacer(modifier = Modifier.width(2.dp))
                        Text(
                            text = "$totalPoints",
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFFFF9800)
                        )
                    }
                }

                Box(
                    modifier = Modifier
                        .width(180.dp)
                        .background(
                            if (canExchange)
                                Brush.linearGradient(
                                    colors = listOf(
                                        Color(0xFFFF6B6B),
                                        Color(0xFFFFA726)
                                    )
                                )
                            else
                                Brush.linearGradient(
                                    colors = listOf(
                                        Color(0xFF9CA3AF),
                                        Color(0xFFD1D5DB)
                                    )
                                ),
                            RoundedCornerShape(28.dp)
                        )
                        .clickable(enabled = canExchange, onClick = onExchangeClick)
                        .padding(vertical = 14.dp),
                    contentAlignment = Alignment.Center
                ) {
                    if (isExchanging) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            androidx.compose.material3.CircularProgressIndicator(
                                color = Color.White,
                                strokeWidth = 2.dp,
                                modifier = Modifier.size(18.dp)
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = "兑换中...",
                                color = Color.White,
                                style = MaterialTheme.typography.titleMedium,
                                fontWeight = FontWeight.SemiBold
                            )
                        }
                    } else {
                        Text(
                            text = if (canExchange) "立即兑换" else "积分不足",
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun QuantitySelector(
    quantity: Int,
    minQuantity: Int,
    maxQuantity: Int,
    onQuantityChange: (Int) -> Unit
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Box(
            modifier = Modifier
                .size(32.dp)
                .background(
                    if (quantity > minQuantity)
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
                    else
                        MaterialTheme.colorScheme.surfaceVariant,
                    RoundedCornerShape(16.dp)
                )
                .clickable(enabled = quantity > minQuantity) {
                    onQuantityChange(quantity - 1)
                },
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "-",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = if (quantity > minQuantity)
                    MaterialTheme.colorScheme.primary
                else
                    MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        Text(
            text = quantity.toString(),
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.width(24.dp),
            textAlign = androidx.compose.ui.text.style.TextAlign.Center
        )

        Box(
            modifier = Modifier
                .size(32.dp)
                .background(
                    if (quantity < maxQuantity)
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.1f)
                    else
                        MaterialTheme.colorScheme.surfaceVariant,
                    RoundedCornerShape(16.dp)
                )
                .clickable(enabled = quantity < maxQuantity) {
                    onQuantityChange(quantity + 1)
                },
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "+",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = if (quantity < maxQuantity)
                    MaterialTheme.colorScheme.primary
                else
                    MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
private fun ExchangeSuccessDialog(
    gift: Gift,
    quantity: Int,
    onDismiss: () -> Unit
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black.copy(alpha = 0.5f))
            .clickable(onClick = onDismiss),
        contentAlignment = Alignment.Center
    ) {
        Box(
            modifier = Modifier
                .width(280.dp)
                .background(Color.White, RoundedCornerShape(20.dp))
                .padding(24.dp)
                .clickable { }
        ) {
            Column(
                modifier = Modifier.fillMaxWidth(),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Box(
                    modifier = Modifier
                        .size(80.dp)
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    Color(0xFF22C55E),
                                    Color(0xFF4ADE80)
                                )
                            ),
                            RoundedCornerShape(40.dp)
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text("✓", color = Color.White, fontSize = 40.sp, fontWeight = FontWeight.Bold)
                }

                Text(
                    text = "兑换成功",
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold
                )

                Text(
                    text = "您已成功兑换 ${gift.name} x$quantity",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = androidx.compose.ui.text.style.TextAlign.Center
                )

                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text("⭐", fontSize = 18.sp)
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "扣除 ${gift.points * quantity} 积分",
                        style = MaterialTheme.typography.bodyMedium,
                        color = Color(0xFFFF9800),
                        fontWeight = FontWeight.Medium
                    )
                }

                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .background(
                            Brush.linearGradient(
                                colors = listOf(
                                    Color(0xFF1D6CFF),
                                    Color(0xFF4D8CFF)
                                )
                            ),
                            RoundedCornerShape(24.dp)
                        )
                        .clickable(onClick = onDismiss)
                        .padding(vertical = 12.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "完成",
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }
}

private fun getGradientColors(categoryId: String): List<Color> {
    return when (categoryId) {
        "cat1" -> listOf(Color(0xFFFFCDD2), Color(0xFFEF9A9A))
        "cat2" -> listOf(Color(0xFFD1C4E9), Color(0xFFB39DDB))
        "cat3" -> listOf(Color(0xFFC5CAE9), Color(0xFF9FA8DA))
        "cat4" -> listOf(Color(0xFFB3E5FC), Color(0xFF81D4FA))
        "cat5" -> listOf(Color(0xFFC8E6C9), Color(0xFFA5D6A7))
        else -> listOf(Color(0xFFFFF3E0), Color(0xFFFFCC80))
    }
}
