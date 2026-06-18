import org.jetbrains.kotlin.gradle.plugin.mpp.KotlinNativeTarget

plugins {
    kotlin("multiplatform")
    id("org.jetbrains.compose")
}

kotlin {
    val isArm64 = System.getProperty("os.arch") == "aarch64"

    val hosArm = if (isArm64) {
        linuxArm64("harmonyArm64")
    } else null

    val hosX64 = linuxX64("harmonyX64")

    listOfNotNull(hosArm, hosX64).forEach { target ->
        target.binaries {
            staticLib {
                baseName = "dispute_shared"
                moduleName = "com.dispute.app"
            }
        }
    }

    sourceSets {
        val commonMain by getting {
            dependencies {
                implementation(project(":shared"))
            }
        }

        val harmonyMain by creating {
            dependsOn(commonMain)
            dependencies {
                implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3")
            }
        }

        val harmonyArm64Main by getting {
            dependsOn(harmonyMain)
        }
        val harmonyX64Main by getting {
            dependsOn(harmonyMain)
        }
    }

    targets.withType<KotlinNativeTarget>().configureEach {
        binaries.withType<org.jetbrains.kotlin.gradle.plugin.mpp.StaticLibrary> {
            linkOpts.add("-s")
        }
    }
}

tasks.register<Copy>("prepareHarmony") {
    group = "harmony"
    description = "Prepare HarmonyOS distribution files"
    dependsOn("linkHarmonyArm64ReleaseStatic", "linkHarmonyX64ReleaseStatic")

    val outDir = "$buildDir/harmony/entry/src/main/libs"

    into(outDir)

    from("build/bin/harmonyArm64/releaseStatic/shared") {
        include("*.a")
        rename { name -> "armeabi-v7a/$name".replace("armeabi-v7a/", "arm64-v8a/") }
    }

    from("build/bin/harmonyX64/releaseStatic/shared") {
        include("*.a")
        into("x86_64/")
    }

    doLast {
        val entryDir = file("$buildDir/harmony/entry/src/main/ets/pages")
        entryDir.mkdirs()

        val pagesList = listOf(
            "Home" to "首页",
            "Login" to "登录",
            "RegisterCase" to "纠纷登记",
            "CaseList" to "我的案件",
            "CaseDetail" to "案件详情",
            "Progress" to "进度查询",
            "AIConsult" to "AI咨询",
            "Satisfaction" to "服务评价",
            "Profile" to "个人中心"
        )

        pagesList.forEach { (pageName, _) ->
            val pageDir = file("$entryDir/${pageName.lowercase()}")
            pageDir.mkdirs()
            val etsFile = file("$pageDir/Index.ets")
            etsFile.writeText("""
                // ${pageName}Page - Auto-generated
                import hilog from '@ohos.hilog';
                import router from '@ohos.router';

                @Entry
                @Component
                struct ${pageName}Page {
                  aboutToAppear() {
                    hilog.info(0x0000, 'DisputeApp', '${pageName}Page appeared');
                  }

                  build() {
                    Column() {
                      Text("${pageName}")
                        .fontSize(24)
                        .fontWeight(FontWeight.Bold)
                        .margin({ top: 24 })
                    }
                    .width('100%')
                    .height('100%')
                    .backgroundColor('#F0F7FF')
                  }
                }
            """.trimIndent())
        }

        val moduleJson = file("$buildDir/harmony/entry/src/main/module.json5")
        moduleJson.parentFile.mkdirs()
        moduleJson.writeText("""
            {
              "module": {
                "name": "entry",
                "type": "entry",
                "description": "$string:module_desc",
                "mainElement": "EntryAbility",
                "deviceTypes": [
                  "phone",
                  "tablet",
                  "2in1"
                ],
                "deliveryWithInstall": true,
                "installationFree": false,
                "pages": "$profile:main_pages",
                "abilities": [
                  {
                    "name": "EntryAbility",
                    "srcEntry": "./ets/entryability/EntryAbility.ets",
                    "description": "$string:EntryAbility_desc",
                    "icon": "$media:icon",
                    "label": "$string:EntryAbility_label",
                    "startWindowIcon": "$media:icon",
                    "startWindowBackground": "$color:start_window_background",
                    "exported": true,
                    "skills": [
                      {
                        "entities": [
                          "entity.system.home"
                        ],
                        "actions": [
                          "action.system.home"
                        ]
                      }
                    ]
                  }
                ],
                "requestPermissions": [
                  {
                    "name": "ohos.permission.INTERNET"
                  }
                ]
              }
            }
        """.trimIndent())
    }
}
