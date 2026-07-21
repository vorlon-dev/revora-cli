package com.example.revora_flutter

import android.content.Context
import android.content.SharedPreferences
import android.util.Base64
import io.flutter.embedding.engine.plugins.FlutterPlugin
import io.flutter.plugin.common.MethodCall
import io.flutter.plugin.common.MethodChannel
import kotlinx.coroutines.*
import org.json.JSONObject
import java.io.File
import java.io.FileOutputStream
import java.net.HttpURLConnection
import java.net.URL
import java.security.KeyFactory
import java.security.PublicKey
import java.security.Signature
import java.security.spec.X509EncodedKeySpec

class RevoraFlutterPlugin : FlutterPlugin, MethodChannel.MethodCallHandler {
    private lateinit var channel: MethodChannel
    private lateinit var context: Context
    private var owner = ""
    private var repo = ""
    private var publicKey: PublicKey? = null
    private val scope = CoroutineScope(Dispatchers.IO)

    override fun onAttachedToEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        context = binding.applicationContext
        channel = MethodChannel(binding.binaryMessenger, "revora_flutter")
        channel.setMethodCallHandler(this)
    }

    override fun onMethodCall(call: MethodCall, result: MethodChannel.Result) {
        when (call.method) {
            "initialize" -> {
                owner = call.argument<String>("owner") ?: ""
                repo = call.argument<String>("repo") ?: ""
                val keyStr = call.argument<String>("publicKey") ?: ""
                publicKey = loadPublicKey(keyStr)
                result.success(null)
            }
            "checkForUpdate" -> {
                scope.launch {
                    val updateAvailable = checkForUpdateAsync()
                    withContext(Dispatchers.Main) { result.success(updateAvailable) }
                }
            }
            "applyUpdate" -> {
                scope.launch {
                    val applied = applyUpdateAsync()
                    withContext(Dispatchers.Main) { result.success(applied) }
                }
            }
            "getCurrentVersion" -> {
                result.success(getCurrentVersion())
            }
            else -> result.notImplemented()
        }
    }

    override fun onDetachedFromEngine(binding: FlutterPlugin.FlutterPluginBinding) {
        channel.setMethodCallHandler(null)
    }

    private fun loadPublicKey(keyStr: String): PublicKey? {
        return try {
            val keyBytes = Base64.decode(keyStr, Base64.DEFAULT)
            val spec = X509EncodedKeySpec(keyBytes)
            KeyFactory.getInstance("RSA").generatePublic(spec)
        } catch (e: Exception) { null }
    }

    private suspend fun checkForUpdateAsync(): Boolean {
        val currentVersion = getCurrentVersion()
        val url = URL("https://api.github.com/repos/$owner/$repo/releases/latest")
        return try {
            val conn = url.openConnection() as HttpURLConnection
            conn.connectTimeout = 5000
            conn.readTimeout = 5000
            val json = JSONObject(conn.inputStream.bufferedReader().readText())
            val latestTag = json.optString("tag_name", "0.0.0")
            if (latestTag != currentVersion) {
                val assets = json.getJSONArray("assets")
                for (i in 0 until assets.length()) {
                    val asset = assets.getJSONObject(i)
                    val name = asset.getString("name")
                    val downloadUrl = asset.getString("browser_download_url")
                    when (name) {
                        "update.patch" -> getPrefs().edit().putString("revora_patch_url", downloadUrl).apply()
                        "manifest.json" -> getPrefs().edit().putString("revora_manifest_url", downloadUrl).apply()
                        "manifest.json.sig" -> getPrefs().edit().putString("revora_signature_url", downloadUrl).apply()
                    }
                }
                true
            } else false
        } catch (e: Exception) { false }
    }

    private suspend fun applyUpdateAsync(): Boolean {
        val prefs = getPrefs()
        val manifestUrl = prefs.getString("revora_manifest_url", null) ?: return false
        val patchUrl = prefs.getString("revora_patch_url", null) ?: return false
        val signatureUrl = prefs.getString("revora_signature_url", null) ?: return false

        val manifestData = downloadString(manifestUrl) ?: return false
        val signatureB64 = downloadString(signatureUrl) ?: return false
        val signatureBytes = Base64.decode(signatureB64, Base64.DEFAULT)

        if (publicKey == null) return false
        val sig = Signature.getInstance("SHA256withRSA")
        sig.initVerify(publicKey)
        sig.update(manifestData.toByteArray())
        if (!sig.verify(signatureBytes)) return false

        val patchFile = File(context.cacheDir, "update.patch")
        downloadFile(patchUrl, patchFile) ?: return false

        val pendingDir = File(context.filesDir, "revora_pending")
        pendingDir.mkdirs()
        patchFile.copyTo(File(pendingDir, "update.patch"), overwrite = true)
        prefs.edit().putBoolean("revora_pending_restart", true).apply()
        return true
    }

    private fun downloadString(urlStr: String): String? {
        return try {
            val conn = URL(urlStr).openConnection() as HttpURLConnection
            conn.connectTimeout = 5000
            conn.inputStream.bufferedReader().readText()
        } catch (e: Exception) { null }
    }

    private fun downloadFile(urlStr: String, dest: File): File? {
        return try {
            val conn = URL(urlStr).openConnection() as HttpURLConnection
            conn.connectTimeout = 15000
            conn.readTimeout = 15000
            conn.inputStream.use { input ->
                FileOutputStream(dest).use { output -> input.copyTo(output) }
            }
            dest
        } catch (e: Exception) { null }
    }

    private fun getCurrentVersion(): String {
        return try {
            val pkgInfo = context.packageManager.getPackageInfo(context.packageName, 0)
            pkgInfo.versionName ?: "0.0.0"
        } catch (e: Exception) { "0.0.0" }
    }

    private fun getPrefs(): SharedPreferences {
        return context.getSharedPreferences("revora", Context.MODE_PRIVATE)
    }
}