import java.io.IOException
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.time.Duration

object GarbageCollectorClient {
    private val client = HttpClient.newBuilder()
        .connectTimeout(Duration.ofSeconds(5))
        .build()

    @JvmStatic
    @JvmOverloads
    @Throws(IOException::class, InterruptedException::class)
    fun collect(
        endpoint: URI = URI.create("http://127.0.0.1:8765/collect")
    ): String {
        val request = HttpRequest.newBuilder(endpoint)
            .timeout(Duration.ofSeconds(10))
            .POST(HttpRequest.BodyPublishers.noBody())
            .build()
        val response = client.send(request, HttpResponse.BodyHandlers.ofString())
        if (response.statusCode() != 200) {
            throw IOException(
                "Garbage collector returned HTTP ${response.statusCode()}: ${response.body()}"
            )
        }
        return response.body()
    }
}
