package edu.sandiego.comp375.jukebox;

import java.io.BufferedInputStream;
import java.io.Closeable;
import java.io.IOException;

import javax.sound.sampled.AudioFormat;
import javax.sound.sampled.AudioInputStream;
import javax.sound.sampled.AudioSystem;
import javax.sound.sampled.DataLine.Info;
import javax.sound.sampled.LineUnavailableException;
import javax.sound.sampled.SourceDataLine;
import javax.sound.sampled.UnsupportedAudioFileException;
import javax.sound.sampled.LineListener;
import javax.sound.sampled.LineEvent;
 
import static javax.sound.sampled.AudioSystem.getAudioInputStream;
import static javax.sound.sampled.AudioFormat.Encoding.PCM_SIGNED;

public class AudioPlayerThread implements Runnable {
	private BufferedInputStream in;
	private final AudioInputStream ais;
	private final SourceDataLine sdl;
	private final AudioInputStream ais2;

	public AudioPlayerThread(BufferedInputStream is) throws Exception {
		this.in = is;
		ais = AudioSystem.getAudioInputStream(in);
		final AudioFormat outFormat = AudioPlayerThread.getOutFormat(ais.getFormat());
		final Info info = new Info(SourceDataLine.class, outFormat);
		this.sdl = (SourceDataLine)AudioSystem.getLine(info);
		this.ais2 = AudioSystem.getAudioInputStream(outFormat, ais);
		sdl.open(outFormat);
	}

	public void run() {
		try {
			final byte[] dataArray = new byte[64*1024];

			sdl.start();
			for (int n = 0; n != -1; n = ais2.read(dataArray, 0, dataArray.length)) {
				// Spin here until there is enough space to write to audio
				// channel.
				// This will help us avoid blocking when calling sdl's write
				// method.
				while (sdl.available() < n) {
					Thread.sleep(100);
				}

				sdl.write(dataArray, 0, n);

				if (Thread.interrupted()) {
					throw new InterruptedException();
				}
			}
			sdl.stop();
		}
		catch (Exception e) {
			System.out.println(e);
		}
		finally {
			sdl.flush();
			sdl.stop();
			sdl.close();
			AudioPlayerThread.closeQuietly(ais);
			AudioPlayerThread.closeQuietly(ais2);
		}
	}

	static AudioFormat getOutFormat(AudioFormat inFormat) {
		final int ch = inFormat.getChannels();
		final float rate = inFormat.getSampleRate();
		return new AudioFormat(AudioFormat.Encoding.PCM_SIGNED, rate, 16, ch,
									ch * 2, rate, false);
	}

	public static void closeQuietly(Closeable c) {
		try {
			if (c != null) {
				c.close();
			}
		} catch (IOException ioe) {
			// ignore
		}
	}
}
