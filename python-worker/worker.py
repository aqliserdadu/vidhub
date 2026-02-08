#!/usr/bin/env python3
"""
Python worker service for yt-dlp integration
Handles video metadata extraction and downloading
"""

from flask import Flask, request, jsonify, send_file
import yt_dlp
import os
import logging
import json
from functools import wraps
import traceback
from datetime import datetime
import subprocess

# Initialize Flask app
app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 500 * 1024 * 1024  # 500MB max

# Configure logging
log_dir = './log'
os.makedirs(log_dir, exist_ok=True)

logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(os.path.join(log_dir, 'worker.log')),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

# Configuration
DOWNLOAD_DIR = os.getenv('DOWNLOAD_DIR', './downloads')
MAX_VIDEO_SIZE_MB = int(os.getenv('MAX_VIDEO_SIZE_MB', 300))
ALLOWED_DOMAINS = os.getenv('ALLOWED_DOMAINS', 'youtube.com,youtu.be,vimeo.com,facebook.com,m.facebook.com,fb.watch,tiktok.com,instagram.com,twitter.com,x.com').split(',')
MAX_FILENAME_LENGTH = int(os.getenv('MAX_FILENAME_LENGTH', 200))

# Ensure download directory exists
os.makedirs(DOWNLOAD_DIR, exist_ok=True)
os.makedirs('./log', exist_ok=True)


def error_handler(f):
    """Decorator for handling errors in API endpoints"""
    @wraps(f)
    def decorated(*args, **kwargs):
        try:
            return f(*args, **kwargs)
        except Exception as e:
            logger.error(f"Error in {f.__name__}: {str(e)}\n{traceback.format_exc()}")
            return jsonify({
                'error': 'server_error',
                'message': str(e),
                'code': 500
            }), 500
    return decorated


def validate_url(url):
    """Validate if URL is from allowed domain"""
    for domain in ALLOWED_DOMAINS:
        if domain.strip() in url:
            return True
    return False


def truncate_filename(filename, max_length):
    """Truncate filename to max length while preserving extension
    Uses proper Unicode handling to avoid breaking multi-byte characters"""
    if len(filename) <= max_length:
        return filename
    
    # Find the extension
    last_dot = filename.rfind('.')
    if last_dot == -1:
        # No extension, just truncate
        return filename[:max_length]
    
    ext = filename[last_dot:]
    base_name = filename[:last_dot]
    
    # Calculate available space for base name (using character count, not bytes)
    available_len = max_length - len(ext)
    if available_len <= 0:
        # Extension is too long, just truncate everything
        return filename[:max_length]
    
    # Truncate base name and add extension
    return base_name[:available_len] + ext


def get_ydl_options(video_url):
    """Get yt-dlp options based on video source"""
    user_agent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
    
    base_options = {
        'quiet': False,
        'no_warnings': False,
        'extract_flat': False,
        'noplaylist': True,
        'socket_timeout': 30,
        'http_headers': {
            'User-Agent': user_agent,
            'Accept-Language': 'en-US,en;q=0.9'
        }
    }
    
    # Facebook-specific options
    if 'facebook.com' in video_url or 'fb.watch' in video_url or 'm.facebook.com' in video_url:
        base_options.update({
            'http_headers': {
                'User-Agent': user_agent,
                'Accept-Language': 'en-US,en;q=0.9',
                'Referer': 'https://www.facebook.com/'
            },
            'quiet': True,
            'no_warnings': True,
            'format_sort': ['res', 'fps', 'codec:h264', 'lang'],
            'fragment_retries': 3,
            'skip_unavailable_fragments': True,
        })
    
    # TikTok-specific options
    elif 'tiktok.com' in video_url or 'vt.tiktok.com' in video_url or 'm.tiktok.com' in video_url:
        base_options.update({
            'http_headers': {
                'User-Agent': user_agent,
                'Accept-Language': 'en-US,en;q=0.9',
                'Referer': 'https://www.tiktok.com/'
            },
            'quiet': False,
            'no_warnings': False,
            'socket_timeout': 60,
            'retries': 3,
            'skip_unavailable_fragments': True,
            'format_sort': ['res', 'fps']
        })
    
    # Instagram-specific options
    elif 'instagram.com' in video_url or 'instagram.com' in video_url:
        base_options.update({
            'http_headers': {
                'User-Agent': user_agent,
                'Accept-Language': 'en-US,en;q=0.9',
                'Referer': 'https://www.instagram.com/'
            },
            'quiet': False,
            'no_warnings': False,
            'socket_timeout': 60,
            'retries': 3,
            'skip_unavailable_fragments': True,
            'format_sort': ['res', 'fps']
        })
    
    # Twitter/X-specific options
    elif 'twitter.com' in video_url or 'x.com' in video_url:
        base_options.update({
            'http_headers': {
                'User-Agent': user_agent,
                'Accept-Language': 'en-US,en;q=0.9',
                'Referer': 'https://twitter.com/'
            },
            'quiet': False,
            'no_warnings': False,
            'socket_timeout': 60,
            'retries': 3,
            'skip_unavailable_fragments': True,
            'format_sort': ['res', 'fps']
        })
    
    return base_options


@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'service': 'yt-dlp-worker',
        'timestamp': datetime.now().isoformat()
    }), 200


@app.route('/api/info', methods=['POST'])
@error_handler
def get_video_info():
    """Get video information from URL"""
    data = request.get_json()
    
    if not data or 'url' not in data:
        return jsonify({
            'error': 'invalid_request',
            'message': 'URL is required',
            'code': 400
        }), 400
    
    video_url = data['url']
    
    # Validate URL domain
    if not validate_url(video_url):
        logger.warning(f"Domain not allowed: {video_url}")
        return jsonify({
            'error': 'invalid_domain',
            'message': 'Domain is not allowed',
            'code': 400
        }), 400
    
    logger.info(f"Fetching info for URL: {video_url}")
    
    try:
        ydl_opts = get_ydl_options(video_url)
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(video_url, download=False)
            
            # Process formats
            formats = []
            if 'formats' in info:
                for fmt in info['formats']:
                    # Skip storyboard and image formats
                    ext = fmt.get('ext', '')
                    if ext in ('mhtml', 'jpg', 'jpeg', 'png', 'gif', 'webp'):
                        continue
                    
                    # Get video/audio info
                    vcodec = fmt.get('vcodec', 'none')
                    acodec = fmt.get('acodec', 'none')
                    
                    # Skip if only has unknown codecs and no extension  
                    if not ext or (vcodec == 'none' and acodec == 'none'):
                        continue
                    
                    format_info = {
                        'format_id': fmt.get('format_id', ''),
                        'ext': ext,
                        'resolution': fmt.get('resolution', 'unknown'),
                        'vcodec': vcodec,
                        'acodec': acodec,
                        'filesize': fmt.get('filesize', 0),
                        'fps': fmt.get('fps', 0),
                        'format': fmt.get('format', ''),
                    }
                    formats.append(format_info)
            
            # If no formats found, try to fetch them again with different options  
            if not formats:
                logger.warning(f"No formats found for {video_url}, retrying with different options...")
                retry_opts = get_ydl_options(video_url)
                retry_opts['skip_unavailable_fragments'] = False
                with yt_dlp.YoutubeDL(retry_opts) as ydl:
                    info = ydl.extract_info(video_url, download=False)
                    if 'formats' in info:
                        for fmt in info['formats']:
                            if fmt.get('ext') and fmt.get('ext') not in ('mhtml', 'jpg', 'jpeg', 'png', 'gif', 'webp'):
                                format_info = {
                                    'format_id': fmt.get('format_id', ''),
                                    'ext': fmt.get('ext', ''),
                                    'resolution': fmt.get('resolution', 'unknown'),
                                    'vcodec': fmt.get('vcodec', 'none'),
                                    'acodec': fmt.get('acodec', 'none'),
                                    'filesize': fmt.get('filesize', 0),
                                    'fps': fmt.get('fps', 0),
                                    'format': fmt.get('format', ''),
                                }
                                formats.append(format_info)
            
            response = {
                'id': info.get('id', ''),
                'title': info.get('title', 'Unknown'),
                'duration': info.get('duration', 0),
                'thumbnail': info.get('thumbnail', ''),
                'uploader': info.get('uploader', 'Unknown'),
                'url': video_url,
                'formats': formats
            }
            
            logger.info(f"Successfully fetched info. Formats: {len(formats)}")
            return jsonify(response), 200
            
    except Exception as e:
        logger.error(f"Failed to fetch video info: {str(e)}")
        return jsonify({
            'error': 'fetch_failed',
            'message': f"Failed to fetch video information: {str(e)}",
            'code': 400
        }), 400


def get_format_with_audio(base_format_id, video_url):
    """
    Construct format string to ensure audio is included
    For video-only formats, merge with best audio: "format_id+bestaudio"
    """
    # First, try to get format info to check if it has audio
    try:
        ydl_opts = get_ydl_options(video_url)
        ydl_opts['skip_download'] = True
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(video_url, download=False)
            
            # Find the selected format
            if 'formats' in info:
                for fmt in info['formats']:
                    if fmt.get('format_id') == base_format_id:
                        acodec = fmt.get('acodec', 'none')
                        vcodec = fmt.get('vcodec', 'none')
                        
                        # If format has no audio but has video, merge with audio
                        if acodec == 'none' and vcodec != 'none':
                            logger.info(f"Format {base_format_id} has no audio, merging with best audio")
                            return f"{base_format_id}+bestaudio/best"
                        # If format has audio, use it as-is
                        elif acodec != 'none':
                            logger.info(f"Format {base_format_id} has audio, using as-is")
                            return base_format_id
                        # If format is audio-only or unknown
                        else:
                            logger.info(f"Format {base_format_id} is audio-only or unknown, using as-is")
                            return base_format_id
            
            # If format not found in list, try merging anyway
            logger.warning(f"Could not determine format {base_format_id} type, attempting merge with audio")
            return f"{base_format_id}+bestaudio/best"
    except Exception as e:
        logger.warning(f"Error checking format type: {str(e)}, using format as-is")
        return base_format_id


@app.route('/api/download', methods=['POST'])
@error_handler
def download_video():
    """Download video with specified format"""
    data = request.get_json()
    
    if not data or 'url' not in data or 'format_id' not in data:
        return jsonify({
            'error': 'invalid_request',
            'message': 'URL and format_id are required',
            'code': 400
        }), 400
    
    video_url = data['url']
    format_id = data['format_id']
    quality = data.get('quality', 'Unknown')  # Get quality label from request
    
    # Validate URL
    if not validate_url(video_url):
        logger.warning(f"Domain not allowed for download: {video_url}")
        return jsonify({
            'error': 'invalid_domain',
            'message': 'Domain is not allowed',
            'code': 400
        }), 400
    
    logger.info(f"Starting download. URL: {video_url}, Format: {format_id}, Quality: {quality}")
    
    try:
        ydl_opts = get_ydl_options(video_url)
        
        # Get format with audio merging if needed
        format_spec = get_format_with_audio(format_id, video_url)
        
        # Determine if this is a merge operation
        is_merge = '+' in format_spec
        
        # Get format info to determine target extension
        target_ext = 'mp4'  # default
        try:
            ydl_test = get_ydl_options(video_url)
            ydl_test['skip_download'] = True
            with yt_dlp.YoutubeDL(ydl_test) as ydl:
                info = ydl.extract_info(video_url, download=False)
                if 'formats' in info:
                    for fmt in info['formats']:
                        if fmt.get('format_id') == format_id:
                            target_ext = fmt.get('ext', 'mp4')
                            break
        except:
            pass
        
        # Set quality suffix for filename
        quality_suffix = f"_{quality}" if quality and quality != 'Unknown' else ""
        
        ydl_opts.update({
            'format': format_spec,
            'outtmpl': os.path.join(DOWNLOAD_DIR, '%(title)s.%(ext)s'),  # No quality suffix in template
            'socket_timeout': 60,
            'noplaylist': True,
            'postprocessors': [],
        })
        
        # Download the video
        logger.debug(f"Starting yt-dlp download with format: {format_spec}")
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            try:
                info = ydl.extract_info(video_url, download=True)
            except Exception as format_error:
                # If format fails (common with Facebook), try best format
                logger.warning(f"Format {format_spec} failed, retrying with best available format: {str(format_error)}")
                ydl_opts['format'] = 'best'  # Fallback to best available
                info = ydl.extract_info(video_url, download=True)
        
        # Get the downloaded filename from yt-dlp
        filename = ydl.prepare_filename(info)
        filepath = os.path.join(DOWNLOAD_DIR, filename)
        
        # Verify file exists after yt-dlp download
        if not os.path.exists(filepath):
            logger.error(f"Downloaded file not found: {filepath}")
            return jsonify({
                'error': 'download_failed',
                'message': 'File was not created during download',
                'code': 400
            }), 400
        
        # Convert merged format if needed
        if is_merge and not filename.lower().endswith(f'.{target_ext}'):
            logger.info(f"Converting merged output to .{target_ext}")
            new_filename = filename.rsplit('.', 1)[0] + f'.{target_ext}'
            new_filepath = os.path.join(DOWNLOAD_DIR, new_filename)
            
            try:
                subprocess.run([
                    'ffmpeg', '-i', filepath, '-c', 'copy', '-y', new_filepath
                ], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                os.remove(filepath)
                filepath = new_filepath
                filename = new_filename
                logger.info(f"Conversion successful: {new_filename}")
            except subprocess.CalledProcessError as e:
                logger.warning(f"Conversion failed, keeping original: {str(e)}")
        
        # Now handle truncation and quality suffix
        # Pre-truncate to account for quality suffix
        if quality_suffix:
            truncate_length = MAX_FILENAME_LENGTH - len(quality_suffix)
            filename = truncate_filename(filename, truncate_length)
            logger.debug(f"Pre-truncated filename to {truncate_length} chars")
        else:
            filename = truncate_filename(filename, MAX_FILENAME_LENGTH)
            logger.debug(f"Truncated filename to {MAX_FILENAME_LENGTH} chars")
        
        # Add quality suffix if needed (single operation, no truncation)
        if quality_suffix:
            base_name, ext = os.path.splitext(filename)
            new_filename = f"{base_name}{quality_suffix}{ext}"
            new_filepath = os.path.join(DOWNLOAD_DIR, new_filename)
            
            try:
                os.rename(filepath, new_filepath)
                filepath = new_filepath
                filename = new_filename
                logger.info(f"Renamed with quality suffix: {new_filename} (length: {len(new_filename)})")
            except Exception as e:
                logger.warning(f"Failed to rename: {str(e)}, keeping original filename")
                # If rename fails, keep the original filepath and filename
        
        # Verify final file exists
        if not os.path.exists(filepath):
            logger.error(f"Final file not found: {filepath}")
            return jsonify({
                'error': 'download_failed',
                'message': 'File was not found after processing',
                'code': 400
            }), 400
        
        # Check file size
        file_size = os.path.getsize(filepath)
        if file_size > MAX_VIDEO_SIZE_MB * 1024 * 1024:
            os.remove(filepath)
            logger.warning(f"File size exceeds limit: {file_size} bytes")
            return jsonify({
                'error': 'file_too_large',
                'message': f'File size exceeds maximum limit of {MAX_VIDEO_SIZE_MB}MB',
                'code': 400
            }), 400
        
        # Extract just the basename to send to Go backend (no directory path)
        download_filename = os.path.basename(filepath)
        
        logger.info(f"Download completed. File: {filepath}, Size: {file_size} bytes, Sending as: {download_filename}")
        
        # Send file to Golang backend with ONLY filename (no path)
        return send_file(
            filepath,
            as_attachment=True,
            download_name=download_filename,  # ONLY basename, no path!
            mimetype='application/octet-stream'
        )
            
    except Exception as e:
        logger.error(f"Download failed: {str(e)}")
        return jsonify({
            'error': 'download_failed',
            'message': f"Download failed: {str(e)}",
            'code': 400
        }), 400
            
    except Exception as e:
        logger.error(f"Download failed: {str(e)}")
        return jsonify({
            'error': 'download_failed',
            'message': f"Download failed: {str(e)}",
            'code': 400
        }), 400


@app.errorhandler(413)
def request_entity_too_large(error):
    """Handle file too large error"""
    logger.warning("Request entity too large")
    return jsonify({
        'error': 'request_too_large',
        'message': f'File size exceeds maximum limit of {MAX_VIDEO_SIZE_MB}MB',
        'code': 413
    }), 413


@app.errorhandler(500)
def internal_error(error):
    """Handle internal server error"""
    logger.error(f"Internal server error: {error}")
    return jsonify({
        'error': 'server_error',
        'message': 'Internal server error',
        'code': 500
    }), 500


if __name__ == '__main__':
    port = int(os.getenv('PYTHON_WORKER_PORT', 5000))
    host = os.getenv('PYTHON_WORKER_HOST', '0.0.0.0')
    
    logger.info(f"Starting Python worker on {host}:{port}")
    app.run(host=host, port=port, debug=False, threaded=True)
