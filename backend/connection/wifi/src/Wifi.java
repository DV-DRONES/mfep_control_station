import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.EOFException;
import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.SocketTimeoutException;
import java.nio.charset.StandardCharsets;
import java.nio.file.AtomicMoveNotSupportedException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.Locale;

// TCP .yaml and .json transciver / reciver
public final class Wifi {
    public static final int DEFAULT_PORT = 50000;
    public static final long DEFAULT_MAX_BYTES = 100L * 1024 * 1024;
    private static final int BUFFER_SIZE = 64 * 1024;
    private static final int MAX_FILE_NAME_BYTES = 1024;

    private Wifi() {
    }

    public static void main(String[] args) throws IOException {
        if (args.length >= 1 && args[0].equalsIgnoreCase("receive")) {
            int port = args.length >= 2 ? Integer.parseInt(args[1]) : DEFAULT_PORT;
            Path directory = args.length >= 3 ? Path.of(args[2]) : Path.of("received");
            System.out.println("Received: " + receive(port, directory));
            return;
        }

        if (args.length >= 3 && args[0].equalsIgnoreCase("send")) {
            int port = args.length >= 4 ? Integer.parseInt(args[3]) : DEFAULT_PORT;
            send(args[1], Path.of(args[2]), port, 10_000);
            System.out.println("File sent.");
            return;
        }

        printUsage();
    }

    private static void printUsage() {
        System.out.println("Wi-Fi file transfer");
        System.out.println("  java Wifi receive [port] [save-directory]");
        System.out.println("  java Wifi send <receiver-ip> <file-path> [port]");
    }

    public static void send(String ip, String filePath) throws IOException {
        send(ip, Path.of(filePath));
    }

    public static void send(String ip, Path file) throws IOException {
        send(ip, file, DEFAULT_PORT, 10_000);
    }

    public static void send(String ip, Path file, int port, int timeoutMillis) throws IOException {
        validatePort(port);
        validateTimeout(timeoutMillis);
        Path source = file.toAbsolutePath().normalize();
        String fileName = source.getFileName().toString();
        validateFileName(fileName);
        if (!Files.isRegularFile(source)) {
            throw new IOException("File does not exist or is not a regular file: " + source);
        }

        byte[] nameBytes = fileName.getBytes(StandardCharsets.UTF_8);
        long fileSize = Files.size(source);

        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress(ip, port), timeoutMillis);
            socket.setSoTimeout(timeoutMillis);
            try (DataOutputStream output = new DataOutputStream(
                    new BufferedOutputStream(socket.getOutputStream()));
                 var input = Files.newInputStream(source)) {
                output.writeInt(nameBytes.length);
                output.write(nameBytes);
                output.writeLong(fileSize);

                byte[] buffer = new byte[BUFFER_SIZE];
                long remaining = fileSize;
                while (remaining > 0) {
                    int count = input.read(buffer, 0, (int) Math.min(buffer.length, remaining));
                    if (count < 0) {
                        throw new EOFException("The file changed while it was being sent");
                    }
                    output.write(buffer, 0, count);
                    remaining -= count;
                }
                output.flush();
            }
        }
    }

    public static Path receive() throws IOException {
        return receive(DEFAULT_PORT, Path.of("received"));
    }

    public static Path receive(int port, Path saveDirectory) throws IOException {
        return receive("0.0.0.0", port, saveDirectory, 0, DEFAULT_MAX_BYTES);
    }

    
    public static Path receive(
            String bindAddress,
            int port,
            Path saveDirectory,
            int timeoutMillis,
            long maxBytes) throws IOException {
        validatePort(port);
        validateTimeout(timeoutMillis);
        if (maxBytes < 0) {
            throw new IllegalArgumentException("maxBytes must not be negative");
        }

        Path directory = saveDirectory.toAbsolutePath().normalize();
        Files.createDirectories(directory);
        Path temporaryFile;

        try (ServerSocket server = new ServerSocket()) {
            server.setReuseAddress(true);
            server.bind(new InetSocketAddress(bindAddress, port));
            server.setSoTimeout(timeoutMillis);

            try (Socket socket = server.accept()) {
                socket.setSoTimeout(timeoutMillis);
                try (DataInputStream input = new DataInputStream(
                        new BufferedInputStream(socket.getInputStream()))) {
                    int nameLength = input.readInt();
                    if (nameLength <= 0 || nameLength > MAX_FILE_NAME_BYTES) {
                        throw new IOException("Invalid file name length");
                    }
                    byte[] nameBytes = new byte[nameLength];
                    input.readFully(nameBytes);
                    String fileName = new String(nameBytes, StandardCharsets.UTF_8);
                    validateFileName(fileName);

                    long fileSize = input.readLong();
                    if (fileSize < 0 || fileSize > maxBytes) {
                        throw new IOException("File size is invalid or exceeds the receive limit");
                    }

                    temporaryFile = Files.createTempFile(directory, ".wifi-", ".part");
                    try {
                        copyExactly(input, temporaryFile, fileSize);
                        Path destination = directory.resolve(fileName);
                        try {
                            Files.move(
                                    temporaryFile,
                                    destination,
                                    StandardCopyOption.ATOMIC_MOVE,
                                    StandardCopyOption.REPLACE_EXISTING);
                        } catch (AtomicMoveNotSupportedException exception) {
                            Files.move(
                                    temporaryFile,
                                    destination,
                                    StandardCopyOption.REPLACE_EXISTING);
                        }
                        return destination;
                    } finally {
                        Files.deleteIfExists(temporaryFile);
                    }
                }
            } catch (SocketTimeoutException exception) {
                throw new SocketTimeoutException("Timed out while waiting for a file");
            }
        }
    }

    /** Compatibility spelling; prefer {@link #receive()} or {@link #receive(int, Path)}. */
    public static Path recive() throws IOException {
        return receive();
    }

    /** Compatibility spelling; prefer {@link #receive(int, Path)}. */
    public static Path recive(int port, Path saveDirectory) throws IOException {
        return receive(port, saveDirectory);
    }

    private static void copyExactly(DataInputStream input, Path destination, long fileSize)
            throws IOException {
        try (var output = new BufferedOutputStream(Files.newOutputStream(destination))) {
            byte[] buffer = new byte[BUFFER_SIZE];
            long remaining = fileSize;
            while (remaining > 0) {
                int count = input.read(buffer, 0, (int) Math.min(buffer.length, remaining));
                if (count < 0) {
                    throw new EOFException("Connection closed before the file was fully received");
                }
                output.write(buffer, 0, count);
                remaining -= count;
            }
        }
    }

    private static void validateFileName(String fileName) {
        String lowerName = fileName.toLowerCase(Locale.ROOT);
        boolean supportedExtension = lowerName.endsWith(".yaml")
                || lowerName.endsWith(".yml")
                || lowerName.endsWith(".json")
                || lowerName.endsWith(".jso");
        if (fileName.isBlank()
                || fileName.equals(".")
                || fileName.equals("..")
                || fileName.contains("/")
                || fileName.contains("\\")
                || !supportedExtension) {
            throw new IllegalArgumentException(
                    "Only .yaml, .yml, .json, and .jso file names are supported");
        }
    }

    private static void validatePort(int port) {
        if (port < 1 || port > 65535) {
            throw new IllegalArgumentException("Port must be between 1 and 65535");
        }
    }

    private static void validateTimeout(int timeoutMillis) {
        if (timeoutMillis < 0) {
            throw new IllegalArgumentException("Timeout must not be negative");
        }
    }
}
