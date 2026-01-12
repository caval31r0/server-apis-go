<?php
header('Content-Type: application/json; charset=utf-8');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST');

// Verifica se o CNPJ foi fornecido via GET ou POST
$cnpj = isset($_POST['cnpj']) ? $_POST['cnpj'] : (isset($_GET['cnpj']) ? $_GET['cnpj'] : '');
$cnpj = preg_replace('/[^0-9]/', '', $cnpj);

// Valida o CNPJ
if (empty($cnpj) || strlen($cnpj) !== 14) {
    http_response_code(400);
    echo json_encode([
        'erro' => 'CNPJ inválido',
        'status' => 400
    ]);
    exit;
}

// URL da API real
$api_url = "https://minhareceita.org/{$cnpj}";

// Inicializa o CURL
$ch = curl_init();

// Configura as opções do CURL
curl_setopt($ch, CURLOPT_URL, $api_url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
curl_setopt($ch, CURLOPT_HTTPHEADER, array('Content-Type: application/json'));

// Executa a requisição
$response = curl_exec($ch);

// Verifica se houve erro
if(curl_errno($ch)) {
    http_response_code(500);
    echo json_encode([
        'erro' => 'Erro ao consultar a API: ' . curl_error($ch),
        'status' => 500
    ]);
    exit;
}

// Fecha a conexão CURL
curl_close($ch);

// Decodifica a resposta
$data = json_decode($response, true);

// Verifica se a resposta é válida
if (!$data || !isset($data['razao_social'])) {
    http_response_code(404);
    echo json_encode([
        'erro' => 'CNPJ não encontrado ou resposta inválida',
        'status' => 404
    ]);
    exit;
}

// Retorna a resposta da API
echo $response;
?> 