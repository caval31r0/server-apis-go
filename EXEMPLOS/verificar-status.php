<?php
/**
 * Verificar Status do Pagamento PIX
 * 
 * Endpoint para verificar o status de uma transação PIX no banco SQLite local
 */

header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');

ini_set('display_errors', 0);
ini_set('log_errors', 1);
error_reporting(E_ALL);

try {
    // Recebe o ID da transação
    $transaction_id = $_GET['transaction_id'] ?? null;
    
    if (empty($transaction_id)) {
        echo json_encode([
            'success' => false,
            'message' => 'ID da transação não fornecido'
        ]);
        exit;
    }
    
    // Conecta ao SQLite
    $dbPath = __DIR__ . '/database.sqlite';
    
    if (!file_exists($dbPath)) {
        echo json_encode([
            'success' => false,
            'message' => 'Banco de dados não encontrado'
        ]);
        exit;
    }
    
    $db = new PDO("sqlite:$dbPath");
    $db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    // Busca a transação
    $stmt = $db->prepare("SELECT transaction_id, status, valor, nome, email, cpf, pix_code, created_at, updated_at 
                          FROM pedidos 
                          WHERE transaction_id = :transaction_id");
    $stmt->execute(['transaction_id' => $transaction_id]);
    $pedido = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if (!$pedido) {
        echo json_encode([
            'success' => false,
            'message' => 'Transação não encontrada'
        ]);
        exit;
    }
    
    // Retorna os dados da transação
    echo json_encode([
        'success' => true,
        'transaction_id' => $pedido['transaction_id'],
        'status' => $pedido['status'],
        'valor' => $pedido['valor'],
        'nome' => $pedido['nome'],
        'email' => $pedido['email'],
        'cpf' => $pedido['cpf'],
        'pix_code' => $pedido['pix_code'],
        'created_at' => $pedido['created_at'],
        'updated_at' => $pedido['updated_at']
    ]);
    
} catch (Exception $e) {
    echo json_encode([
        'success' => false,
        'message' => 'Erro ao verificar status: ' . $e->getMessage()
    ]);
}
?>
