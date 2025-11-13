"""
示例通知插件
"""

def on_event(event):
    """
    处理事件
    
    Args:
        event (dict): 事件数据
        
    Returns:
        dict: 处理结果
    """
    # 根据事件类型进行不同的处理
    if event['type'] == 'download.complete':
        return handle_download_complete(event)
    elif event['type'] == 'transfer.complete':
        return handle_transfer_complete(event)
    elif event['type'] == 'subscribe.added':
        return handle_subscribe_added(event)
    else:
        return {'status': 'ignored', 'message': 'Unknown event type'}

def handle_download_complete(event):
    """处理下载完成事件"""
    # TODO: 实现下载完成通知逻辑
    return {'status': 'success', 'message': 'Download complete notification sent'}

def handle_transfer_complete(event):
    """处理转移完成事件"""
    # TODO: 实现转移完成通知逻辑
    return {'status': 'success', 'message': 'Transfer complete notification sent'}

def handle_subscribe_added(event):
    """处理订阅添加事件"""
    # TODO: 实现订阅添加通知逻辑
    return {'status': 'success', 'message': 'Subscribe added notification sent'}