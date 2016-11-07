package edu.sandiego.comp375.jukebox;

import java.io.File;

/**
 * A utility class for our Jukebox.
 *
 * DO NOT MODIFY THIS FILE!
 */
public class AudioUtil {
	public static File getSoundFile(String fileName) {
		File soundFile = new File(fileName);
		if (!soundFile.exists() || !soundFile.isFile())
			throw new IllegalArgumentException("not a file: " + soundFile);
		return soundFile;
	}
}
