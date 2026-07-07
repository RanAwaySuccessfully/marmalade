#ifndef LIBSHARED_H
#define LIBSHARED_H

char* ui_getembed(int);
void srv_config();
int srv_status();
char* srv_error();
void srv_start();
void srv_stop();

#endif // LIBSHARED_H