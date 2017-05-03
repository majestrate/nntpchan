/*
 Author: José Bollo <jobol@nonadev.net>
 Author: José Bollo <jose.bollo@iot.bzh>

 https://gitlab.com/jobol/mustach

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

#ifndef _mustach_h_included_
#define _mustach_h_included_

/**
 * mustach_itf - interface for callbacks
 *
 * All of this function should return a negative value to stop
 * the mustache processing. The returned negative value will be
 * then returned to the caller of mustach as is.
 *
 * The functions enter and next should return 0 or 1.
 *
 * All other functions should normally return 0.
 *
 * @start: Starts the mustach processing of the closure
 *         'start' is optional (can be NULL)
 *
 * @put: Writes the value of 'name' to 'file' with 'escape' or not
 *
 * @enter: Enters the section of 'name' if possible.
 *         Musts return 1 if entered or 0 if not entered.
 *         When 1 is returned, the function 'leave' will always be called.
 *         Conversely 'leave' is never called when enter returns 0 or
 *         a negative value.
 *         When 1 is returned, the function must activate the first
 *         item of the section.
 *
 * @next: Activates the next item of the section if it exists.
 *        Musts return 1 when the next item is activated.
 *        Musts return 0 when there is no item to activate.
 *
 * @leave: Leaves the last entered section
 */
struct mustach_itf {
	int (*start)(void *closure);
	int (*put)(void *closure, const char *name, int escape, FILE *file);
	int (*enter)(void *closure, const char *name);
	int (*next)(void *closure);
	int (*leave)(void *closure);
};

#define MUSTACH_OK                       0
#define MUSTACH_ERROR_SYSTEM            -1
#define MUSTACH_ERROR_UNEXPECTED_END    -2
#define MUSTACH_ERROR_EMPTY_TAG         -3
#define MUSTACH_ERROR_TAG_TOO_LONG      -4
#define MUSTACH_ERROR_BAD_SEPARATORS    -5
#define MUSTACH_ERROR_TOO_DEPTH         -6
#define MUSTACH_ERROR_CLOSING           -7
#define MUSTACH_ERROR_BAD_UNESCAPE_TAG  -8

/**
 * fmustach - Renders the mustache 'template' in 'file' for 'itf' and 'closure'.
 *
 * @template: the template string to instanciate
 * @itf:      the interface to the functions that mustach calls
 * @closure:  the closure to pass to functions called
 * @file:     the file where to write the result
 *
 * Returns 0 in case of success, -1 with errno set in case of system error
 * a other negative value in case of error.
 */
extern int fmustach(const char *template, struct mustach_itf *itf, void *closure, FILE *file);

/**
 * fmustach - Renders the mustache 'template' in 'fd' for 'itf' and 'closure'.
 *
 * @template: the template string to instanciate
 * @itf:      the interface to the functions that mustach calls
 * @closure:  the closure to pass to functions called
 * @fd:       the file descriptor number where to write the result
 *
 * Returns 0 in case of success, -1 with errno set in case of system error
 * a other negative value in case of error.
 */
extern int fdmustach(const char *template, struct mustach_itf *itf, void *closure, int fd);

/**
 * fmustach - Renders the mustache 'template' in 'result' for 'itf' and 'closure'.
 *
 * @template: the template string to instanciate
 * @itf:      the interface to the functions that mustach calls
 * @closure:  the closure to pass to functions called
 * @result:   the pointer receiving the result when 0 is returned
 * @size:     the size of the returned result
 *
 * Returns 0 in case of success, -1 with errno set in case of system error
 * a other negative value in case of error.
 */
extern int mustach(const char *template, struct mustach_itf *itf, void *closure, char **result, size_t *size);

#endif

