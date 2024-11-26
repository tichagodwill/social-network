import type { FilePreview, FileUploadResponse } from '$lib/types';

export async function createFilePreview(file: File): Promise<FilePreview> {
    const preview = await new Promise<string>((resolve) => {
        if (file.type.startsWith('image/')) {
            const reader = new FileReader();
            reader.onloadend = () => resolve(reader.result as string);
            reader.readAsDataURL(file);
        } else {
            resolve(URL.createObjectURL(file));
        }
    });

    let type: FilePreview['type'] = 'document';
    if (file.type.startsWith('image/')) {
        type = 'image';
    } else if (file.type.startsWith('video/')) {
        type = 'video';
    } else if (file.type.startsWith('audio/')) {
        type = 'audio';
    }

    return {
        file,
        preview,
        type
    };
}

export async function uploadFile(file: File): Promise<FileUploadResponse> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await fetch('http://localhost:8080/upload', {
        method: 'POST',
        credentials: 'include',
        body: formData
    });

    if (!response.ok) {
        throw new Error('Upload failed');
    }

    return await response.json();
} 