import org.jetbrains.kotlin.gradle.targets.js.dsl.ExperimentalWasmDsl

plugins {
    kotlin("multiplatform")
    id("org.jetbrains.compose")
}

kotlin {
    @OptIn(ExperimentalWasmDsl::class)
    wasmJs {
        browser()
        binaries.executable()
    }

    sourceSets {
        val commonMain by getting {
            dependencies {
                implementation(project(":shared"))
            }
        }
        val wasmJsMain by getting {
            dependencies {
                implementation(compose.html.core)
            }
        }
    }
}

compose.experimental {
    web.application {
        outputDirectory.set(file("$buildDir/dist"))
    }
}

tasks.register<Copy>("copyAlipayResources") {
    from("$buildDir/dist")
    into("$buildDir/alipay-miniprogram")
    filter { line ->
        line
            .replace("%PLATFORM%", "alipay")
            .replace("%APP_NAME%", "纠纷调解")
    }
}

tasks.register("buildAlipayMiniProgram") {
    group = "miniapp"
    description = "Build Alipay mini-program distribution"
    dependsOn("wasmJsBrowserDistribution", "copyAlipayResources")

    doLast {
        val miniAppJson = file("$buildDir/alipay-miniprogram/app.json")
        miniAppJson.parentFile.mkdirs()
        miniAppJson.writeText("""
            {
                "pages": [
                    "pages/home/index",
                    "pages/login/index",
                    "pages/register/index",
                    "pages/caselist/index",
                    "pages/casedetail/index",
                    "pages/progress/index",
                    "pages/aiconsult/index",
                    "pages/satisfaction/index",
                    "pages/profile/index"
                ],
                "window": {
                    "defaultTitle": "纠纷调解服务",
                    "titleBarColor": "#1D6CFF",
                    "titleBarTextColor": "#FFFFFF",
                    "backgroundColor": "#F0F7FF",
                    "allowPullDownRefresh": true
                },
                "tabBar": {
                    "textColor": "#9CA3AF",
                    "selectedColor": "#1D6CFF",
                    "backgroundColor": "#FFFFFF",
                    "items": [
                        {
                            "page": "pages/home/index",
                            "name": "首页",
                            "icon": "assets/tab/home.png",
                            "activeIcon": "assets/tab/home-active.png"
                        },
                        {
                            "page": "pages/caselist/index",
                            "name": "案件",
                            "icon": "assets/tab/case.png",
                            "activeIcon": "assets/tab/case-active.png"
                        },
                        {
                            "page": "pages/aiconsult/index",
                            "name": "咨询",
                            "icon": "assets/tab/ai.png",
                            "activeIcon": "assets/tab/ai-active.png"
                        },
                        {
                            "page": "pages/profile/index",
                            "name": "我的",
                            "icon": "assets/tab/profile.png",
                            "activeIcon": "assets/tab/profile-active.png"
                        }
                    ]
                }
            }
        """.trimIndent())

        val projectConfig = file("$buildDir/alipay-miniprogram/project.config.json")
        projectConfig.writeText("""
            {
                "miniprogramRoot": "./",
                "appid": "2024000000000001",
                "projectname": "dispute-resolve-alipay",
                "description": "纠纷多元化解支付宝小程序",
                "compileType": "mini",
                "compileOptions": {
                    "component2": true,
                    "typescript": true,
                    "minified": true
                },
                "libVersion": "2.0.0"
            }
        """.trimIndent())
    }
}
