import org.jetbrains.kotlin.gradle.targets.js.dsl.ExperimentalWasmDsl

plugins {
    kotlin("multiplatform")
    id("org.jetbrains.compose")
    id("com.google.devtools.ksp")
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

dependencies {
    add("kspMetadata", "com.arkivanov.decompose:extensions-ksp:2.2.0")
}

tasks.register<Copy>("copyWechatResources") {
    from("$buildDir/dist")
    into("$buildDir/wechat-miniprogram")
    filter { line ->
        line
            .replace("%PLATFORM%", "wechat")
            .replace("%APP_NAME%", "纠纷调解")
    }
}

tasks.register("buildWechatMiniProgram") {
    group = "miniapp"
    description = "Build WeChat mini-program distribution"
    dependsOn("wasmJsBrowserDistribution", "copyWechatResources")

    doLast {
        val configFile = file("$buildDir/wechat-miniprogram/project.config.json")
        configFile.parentFile.mkdirs()
        configFile.writeText("""
            {
                "description": "纠纷多元化解服务小程序",
                "packOptions": {
                    "ignore": [],
                    "include": []
                },
                "setting": {
                    "bundle": false,
                    "userConfirmedBundleSwitch": false,
                    "urlCheck": true,
                    "scopeDataCheck": false,
                    "coverView": true,
                    "es6": true,
                    "postcss": true,
                    "compileHotReLoad": false,
                    "lazyloadPlaceholderEnable": false,
                    "preloadBackgroundData": false,
                    "minified": true,
                    "autoAudits": false,
                    "newFeature": false,
                    "uglifyFileName": false,
                    "uploadWithSourceMap": true,
                    "useIsolateContext": true,
                    "nodeModules": false,
                    "enhance": true,
                    "useMultiFrameRuntime": true,
                    "useApiHook": true,
                    "useApiHostProcess": true,
                    "showShadowRootInWxmlPanel": true,
                    "packNpmManually": false,
                    "enableEngineNative": false,
                    "packNpmRelationList": [],
                    "minifyWXSS": true,
                    "showES6CompileOption": false,
                    "minifyWXML": true,
                    "babelSetting": {
                        "ignore": [],
                        "disablePlugins": [],
                        "outputPath": ""
                    }
                },
                "compileType": "miniprogram",
                "libVersion": "3.0.0",
                "appid": "wx_DISPUTE_APP_ID",
                "projectname": "dispute-resolve-miniapp",
                "condition": {},
                "editorSetting": {
                    "tabIndent": "insertSpaces",
                    "tabSize": 2
                }
            }
        """.trimIndent())

        val sitemapFile = file("$buildDir/wechat-miniprogram/sitemap.json")
        sitemapFile.writeText("""
            {
                "desc": "关于本文件的更多信息，请参考文档 https://developers.weixin.qq.com/miniprogram/dev/framework/sitemap.html",
                "rules": [
                    {
                        "action": "allow",
                        "page": "*"
                    }
                ]
            }
        """.trimIndent())
    }
}
