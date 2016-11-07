package edu.sandiego.comp375.jukebox;

import java.io.*;
import java.net.*;

/**
 * This class is a simple server that streams an MP3 file over a socket.
 * It listens on port 6666 and occasionally throws in a 40ms delay between
 * sends (just to make sure the client can handle the occasional network
 * hiccup.
 * You'll need to specify the file you want to stream on the command line.
 *
 * @note This is just to test your initial client. Your real server should be
 * written in C (see the server/ directory).
 *
 * DO NOT MODIFY THIS FILE!
 */
public class AudioServer {
	public static void main(String[] args) throws IOException {
		if (args.length == 0)
			throw new IllegalArgumentException("expected sound file arg");
		File soundFile = AudioUtil.getSoundFile(args[0]);

		System.out.println("server: " + soundFile);

		try {
			ServerSocket serverSocket = new ServerSocket(6666); 
			FileInputStream in = new FileInputStream(soundFile);
			if (serverSocket.isBound()) {
				Socket client = serverSocket.accept();
				OutputStream out = client.getOutputStream();

				byte buffer[] = new byte[2048];
				int count;
				int loopCount = 0;
				while ((count = in.read(buffer)) != -1) {
					out.write(buffer, 0, count);
					loopCount++;
					// Insert a 40ms delay every 10th send, just to show it
					// works if there are gaps in network connectivity.
					if (loopCount % 10 == 0) {
						Thread.sleep(40);
					}
				}

			}
		}
		catch (Exception e) {
			System.out.println(e);
		}

		System.out.println("server: shutdown");
	}
}
