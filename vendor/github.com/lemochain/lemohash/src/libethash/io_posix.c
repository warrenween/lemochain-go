/*
  This file is part of lemohash.

  lemohash is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  lemohash is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with lemohash.  If not, see <http://www.gnu.org/licenses/>.
*/
/** @file io_posix.c
 * @author Lefteris Karapetsas <lefteris@lemodev.com>
 * @date 2015
 */

#include "io.h"
#include <sys/types.h>
#include <sys/stat.h>
#include <errno.h>
#include <libgen.h>
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <pwd.h>

FILE* lemohash_fopen(char const* file_name, char const* mode)
{
	return fopen(file_name, mode);
}

char* lemohash_strncat(char* dest, size_t dest_size, char const* src, size_t count)
{
	return strlen(dest) + count + 1 <= dest_size ? strncat(dest, src, count) : NULL;
}

bool lemohash_mkdir(char const* dirname)
{
	int rc = mkdir(dirname, S_IRWXU | S_IRWXG | S_IROTH | S_IXOTH);
	return rc != -1 || errno == EEXIST;
}

int lemohash_fileno(FILE *f)
{
	return fileno(f);
}

char* lemohash_io_create_filename(
	char const* dirname,
	char const* filename,
	size_t filename_length
)
{
	size_t dirlen = strlen(dirname);
	size_t dest_size = dirlen + filename_length + 1;
	if (dirname[dirlen] != '/') {
		dest_size += 1;
	}
	char* name = malloc(dest_size);
	if (!name) {
		return NULL;
	}

	name[0] = '\0';
	lemohash_strncat(name, dest_size, dirname, dirlen);
	if (dirname[dirlen] != '/') {
		lemohash_strncat(name, dest_size, "/", 1);
	}
	lemohash_strncat(name, dest_size, filename, filename_length);
	return name;
}

bool lemohash_file_size(FILE* f, size_t* ret_size)
{
	struct stat st;
	int fd;
	if ((fd = fileno(f)) == -1 || fstat(fd, &st) != 0) {
		return false;
	}
	*ret_size = st.st_size;
	return true;
}

bool lemohash_get_default_dirname(char* strbuf, size_t buffsize)
{
	static const char dir_suffix[] = ".lemohash/";
	strbuf[0] = '\0';
	char* home_dir = getenv("HOME");
	if (!home_dir || strlen(home_dir) == 0)
	{
		struct passwd* pwd = getpwuid(getuid());
		if (pwd)
			home_dir = pwd->pw_dir;
	}
	
	size_t len = strlen(home_dir);
	if (!lemohash_strncat(strbuf, buffsize, home_dir, len)) {
		return false;
	}
	if (home_dir[len] != '/') {
		if (!lemohash_strncat(strbuf, buffsize, "/", 1)) {
			return false;
		}
	}
	return lemohash_strncat(strbuf, buffsize, dir_suffix, sizeof(dir_suffix));
}
