import pygame
import time

# Initialize pygame mixer
pygame.mixer.init()

flush = "./res/flush_toilet.mp3"
pull = "./res/pull_up.mp3"
wash = "./res/wash_hands.mp3"
well = "./res/well_done.mp3"

# Function to play a single music file
def play_music(music_file):
    pygame.mixer.music.load(music_file)
    pygame.mixer.music.play()
    while pygame.mixer.music.get_busy():  # Wait for the music to finish playing
        time.sleep(1)

# play_music(pull)
playlist = [pull, flush, wash, well]
for audio in playlist:
    play_music(audio)