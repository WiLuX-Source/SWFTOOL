#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>

void show_intro();
FILE *open_input_file(const char *name);
FILE *create_output_file(const char *name);
void fast_data_dump(FILE *input, FILE *output, const size_t length);
void data_dump(FILE *input, FILE *output, const size_t length);
unsigned long int get_file_size(FILE *file);
char *get_string_memory(const size_t length);
size_t get_extension_position(const char *source);
char *get_short_name(const char *name);
char *get_name(const char *name, const char *ext);
void go_offset(FILE *file, const unsigned long int offset);
void check_executable(FILE *input);
void knife(const char *target);
void decompile(const char *target, const char *flash);
unsigned long int get_movie_length(FILE *input);
void compile_flash(const char *player, const char *flash, const char *result);
void magic(const char *player, const char *flash);
void write_service_information(FILE *output, const unsigned long int length);
unsigned long int copy_file(FILE *input, FILE *output);


int main(int argc, char *argv[])
{
  show_intro();
  if (argc == 2 && strcmp(strrchr(argv[1], '.'), ".exe") == 0)
  {
    knife(argv[1]);
    puts("SWF file extracted successfully.");
  }
  else if (argc == 3)
  {
    char *exe;
    char *swf;
    for (size_t i = 1; i < argc; i++)
    {
      if (strcmp(strrchr(argv[i], '.'), ".exe") == 0)
      {
        exe = argv[i];
      }
      if (strcmp(strrchr(argv[i], '.'), ".swf") == 0)
      {
        swf = argv[i];
      }
    }
    magic(exe, swf);
    puts("SWF file combined successfully.");
  }
  else
  {
    puts("You called the program wrong. Please try again.");
  }
}

// Common

void show_intro()
{
  puts("A simple tool for converting SWF -> EXE & EXE -> SWF");
  puts("Built top on magicswf & swfknife");
  puts("Code belongs to Popov Evgeniy Alekseyevich - Github : PopovEvgeniy");
  puts("This software distributed under GNU GENERAL PUBLIC LICENSE");
  puts("");
}

FILE *open_input_file(const char *name)
{
  FILE *target;
  target = fopen(name, "rb");
  if (target == NULL)
  {
    puts("Can't open input file");
  }
  return target;
}

FILE *create_output_file(const char *name)
{
  FILE *target;
  target = fopen(name, "wb");
  if (target == NULL)
  {
    puts("Can't create output file");
  }
  return target;
}

void data_dump(FILE *input, FILE *output, const size_t length)
{
  unsigned char data;
  size_t index;
  data = 0;
  for (index = 0; index < length; ++index)
  {
    fread(&data, sizeof(unsigned char), 1, input);
    fwrite(&data, sizeof(unsigned char), 1, output);
  }
}

void fast_data_dump(FILE *input, FILE *output, const size_t length)
{
  unsigned char *buffer = NULL;
  buffer = (unsigned char *)calloc(length, sizeof(unsigned char));
  if (buffer == NULL)
  {
    data_dump(input, output, length);
  }
  else
  {
    fread(buffer, sizeof(unsigned char), length, input);
    fwrite(buffer, sizeof(unsigned char), length, output);
    free(buffer);
  }
}

unsigned long int get_file_size(FILE *file)
{
  unsigned long int length;
  fseek(file, 0, SEEK_END);
  length = ftell(file);
  rewind(file);
  return length;
}

char *get_string_memory(const size_t length)
{
  char *memory = NULL;
  memory = (char *)calloc(length + 1, sizeof(char));
  if (memory == NULL)
  {
    puts("Can't allocate memory");
  }
  return memory;
}

size_t get_extension_position(const char *source)
{
  size_t index;
  for (index = strlen(source); index > 0; --index)
  {
    if (source[index] == '.')
    {
      break;
    }
  }
  if (index == 0)
    index = strlen(source);
  return index;
}

char *get_short_name(const char *name)
{
  size_t length;
  char *result = NULL;
  length = get_extension_position(name);
  result = get_string_memory(length);
  strncpy(result, name, length);
  return result;
}

char *get_name(const char *name, const char *ext)
{
  char *result = NULL;
  char *output = NULL;
  size_t length;
  output = get_short_name(name);
  length = strlen(output) + strlen(ext);
  result = get_string_memory(length);
  strcpy(result, output);
  free(output);
  return strcat(result, ext);
}

void check_executable(FILE *input)
{
  char signature[2];
  fread(signature, sizeof(char), 2, input);
  if (strncmp(signature, "MZ", 2) != 0)
  {
    puts("Executable file of Flash Player Projector corrupted");
  }
}

// Knife

void check_knife(FILE *input)
{
  unsigned long int signature;
  signature = 0;
  fread(&signature, sizeof(unsigned long int), 1, input);
  if (signature != 0xFA123456)
  {
    puts("Flash movie corrupted");
  }
}

void knife(const char *target)
{
  char *output = NULL;
  output = get_name(target, ".swf");
  decompile(target, output);
  free(output);
}

void decompile(const char *target, const char *flash)
{
  FILE *input;
  FILE *output;
  unsigned long int total, movie;
  input = open_input_file(target);
  check_executable(input);
  total = get_file_size(input);
  go_offset(input, total - 8);
  check_knife(input);
  movie = get_movie_length(input);
  go_offset(input, total - movie - 8);
  output = create_output_file(flash);
  fast_data_dump(input, output, (size_t)movie);
  fclose(input);
  fclose(output);
}

void go_offset(FILE *file, const unsigned long int offset)
{
  if (fseek(file, offset, SEEK_SET) != 0)
  {
    puts("Can't jump to target offset");
  }
}

unsigned long int get_movie_length(FILE *input)
{
  unsigned long int length;
  length = 0;
  fread(&length, sizeof(unsigned long int), 1, input);
  return length;
}



// Magic

void check_magic(FILE *input)
{
  char signature[3];
  fread(signature, sizeof(char), 3, input);
  if (strncmp(signature, "FWS", 3) != 0)
  {
    if (strncmp(signature, "CWS", 3) != 0)
    {
      puts("Flash movie corrupted");
    }
  }
}
void magic(const char *player, const char *flash)
{
  char *output = NULL;
  output = get_name(flash, ".exe");
  compile_flash(player, flash, output);
  free(output);
}

void compile_flash(const char *player, const char *flash, const char *result)
{
  unsigned long int length;
  FILE *projector;
  FILE *swf;
  FILE *output;
  projector = open_input_file(player);
  swf = open_input_file(flash);
  check_executable(projector);
  check_magic(swf);
  output = create_output_file(result);
  copy_file(projector, output);
  length = copy_file(swf, output);
  write_service_information(output, length);
  fclose(projector);
  fclose(swf);
  fclose(output);
}

void write_service_information(FILE *output, const unsigned long int length)
{
  unsigned long int flag;
  flag = 0xFA123456;
  fwrite(&flag, sizeof(unsigned long int), 1, output);
  fwrite(&length, sizeof(unsigned long int), 1, output);
}

unsigned long int copy_file(FILE *input, FILE *output)
{
  unsigned long int length;
  length = get_file_size(input);
  fast_data_dump(input, output, (size_t)length);
  return length;
}