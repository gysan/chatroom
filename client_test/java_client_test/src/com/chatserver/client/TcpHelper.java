package com.chatserver.client;

import java.io.*;
import java.net.Socket;

public class TcpHelper {
	private Socket socket;
	private InputStream in;
	private OutputStream out;
	
	// 连接服务器，成功返回true，失败返回false
	public boolean connect(String ip, int port) {
		try {
			this.socket = new Socket(ip, port);
			this.in = this.socket.getInputStream();
			this.out = this.socket.getOutputStream();
		} catch (Exception e) {
			e.printStackTrace();
			return false;
		}
		return true;
	}
	
	// 关闭连接
	public void close() {
		try {
			this.socket.close();
		} catch (Exception e) {
			e.printStackTrace();
		}
	}
	
	// 发送消息，成功返回true, 失败返回false
	public boolean send(byte[] sendData) {
		try {
			out.write(sendData);
		} catch (Exception e) {
			e.printStackTrace();
			return false;
		}
		return true;
	}

	// 读取消息，成功返回读取字节数，失败返回-1
	public int read(byte[] readData) {
		int readSize = -1;
		try {
			readSize = in.read(readData);
		} catch (Exception e) {
			e.printStackTrace();
			readSize = -1;
		}
		return readSize;
	}
}
