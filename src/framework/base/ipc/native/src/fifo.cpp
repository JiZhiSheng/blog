#include "include/fifo.h"

#include <iostream>
#include <errno.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>

#define BUFFER 4096

int TwoWayFifo::index = 0;

TwoWayFifo::TwoWayFifo(const std::string& name)
    : name_(name), id_(-1), read_fd_(-1), write_fd_(-1) {
    id_ = index++;
}

TwoWayFifo::~TwoWayFifo() {
    Close();
}

bool TwoWayFifo::CreateServerFile() {
    std::string file_path = "/tmp/com.ipc." + name_ + ".server";
    // std::cout << "CreateServerFile: " << file_path << std::endl;
    if (access(file_path.c_str(), F_OK) != 0) {
        if (0 != mkfifo(file_path.c_str(), 0664)) {
            perror("mkfifo error");
        }
    }
    read_fd_ = open(file_path.c_str(), O_RDONLY | O_NONBLOCK);
    if (read_fd_ == -1) {
        perror("CreateServerFile");
        return false;
    }
    return true;
}

bool TwoWayFifo::CreateClientFile() {
    std::string file_path = "/tmp/com.ipc." + name_ + ".client";
    // std::cout << "CreateClientFile: " << file_path << std::endl;
    if (access(file_path.c_str(), F_OK) != 0) {
        if (0 != mkfifo(file_path.c_str(), 0664)) {
            perror("mkfifo error");
        }
    }
    read_fd_ = open(file_path.c_str(), O_RDONLY | O_NONBLOCK);
    if (read_fd_ == -1) {
        perror("CreateClientFile");
        return false;
    }
    return true;
}

bool TwoWayFifo::OpenServerFile() {
    std::string file_path = "/tmp/com.ipc." + name_ + ".server";
    // std::cout << "OpenServerFile: " << file_path << std::endl;
    if (access(file_path.c_str(), F_OK) != 0) {
        return false;
    }
    write_fd_ = open(file_path.c_str(), O_WRONLY | O_NONBLOCK);
    if (write_fd_ == -1) {
        perror("OpenServerFile");
        return false;
    }
    return true;
}

bool TwoWayFifo::OpenClientFile() {
    std::string file_path = "/tmp/com.ipc." + name_ + ".client";
    // std::cout << "OpenClientFile: " << file_path << std::endl;
    if (access(file_path.c_str(), F_OK) != 0) {
        return false;
    }
    write_fd_ = open(file_path.c_str(), O_WRONLY | O_NONBLOCK);
    if (write_fd_ == -1) {
        perror("OpenClientFile");
        return false;
    }
    return true;
}

int TwoWayFifo::Write(const std::string& data) {
    // std::cout << "Write: " << data.size() << std::endl;
    int data_size = data.size();
    std::string real_data = "";
    real_data.append(reinterpret_cast<char*>(&data_size), sizeof(int));
    real_data = real_data + data;
    data_size += sizeof(int);
    const unsigned char* write_buf = reinterpret_cast<const unsigned char*>(real_data.c_str());
    int writed_size = 0;
    while (true) {
        int current_write_size = write(write_fd_, write_buf + writed_size, data_size);
        if ((current_write_size == -1 && errno == EAGAIN) || current_write_size == 0) {
            return 0;
        } else if (current_write_size == -1) {
            perror("write failed");
            return -1;
        }
        writed_size += current_write_size;
        if (writed_size == data_size) {
            return 0;
        }
    }
}

int TwoWayFifo::Read(std::string& data) {
    unsigned char buf[BUFFER];
    while (true) {
        int read_size = read(read_fd_, buf, BUFFER);
        if (read_size == -1 && errno == EAGAIN) {
            return 0;
        } else if (read_size == -1) {
            perror("read failed");
            return -1;
        }
        data.append((const char*)buf, read_size);
    }
}

void TwoWayFifo::Close() {
    if (read_fd_ != -1) {
        close(read_fd_);
    }
    if (write_fd_ != -1) {
        close(write_fd_);
    }
}