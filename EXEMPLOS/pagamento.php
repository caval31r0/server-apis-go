<?php
/*
 * GERADOR DE PIX DINÂMICO
 * 
 * ========================================
 * PARÂMETROS ACEITOS (GET ou POST):
 * ========================================
 * 
 * VALOR:
 * - valor: Valor do pagamento
 *   Exemplos: 2000 (centavos) ou 20.00 (reais)
 *   Padrão: R$ 59,40 (13845 centavos)
 * 
 * DADOS DO CLIENTE:
 * - nome/name: Nome completo do cliente
 *   Exemplo: "João Silva Santos"
 *   Fallback: Nome aleatório gerado
 * 
 * - cpf/document: CPF do cliente (com ou sem formatação)
 *   Exemplos: "12345678900" ou "123.456.789-00"
 *   Fallback: CPF válido gerado
 * 
 * - telefone/telephone/phone: Telefone do cliente
 *   Exemplos: "11999999999" ou "(11) 99999-9999"
 *   Fallback: (11) 99999-9999
 * 
 * - email: Email do cliente
 *   Exemplo: "cliente@email.com"
 *   Fallback: Email gerado baseado no nome
 * 
 * ========================================
 * EXEMPLOS DE USO:
 * ========================================
 * 
 * 1. Via GET (URL):
 *    pagamento.php?valor=2000&nome=João Silva&cpf=12345678900&telefone=11999999999&email=joao@email.com
 * 
 * 2. Via POST (JavaScript):
 *    fetch('pagamento.php', {
 *        method: 'POST',
 *        body: new FormData(document.getElementById('formulario_pagamento'))
 *    });
 * 
 * 3. Via POST (Formulário HTML):
 *    <form action="pagamento.php" method="POST">
 *        <input name="valor" value="20.00">
 *        <input name="nome" value="João Silva">
 *        <input name="cpf" value="123.456.789-00">
 *        <input name="telefone" value="(11) 99999-9999">
 *        <input name="email" value="joao@email.com">
 *        <button type="submit">Gerar PIX</button>
 *    </form>
 * 
 * ========================================
 * FALLBACK AUTOMÁTICO:
 * ========================================
 * Se algum dado não for enviado ou for inválido,
 * o sistema gera automaticamente dados válidos.
 * 
 * Isso garante que o PIX sempre será gerado,
 * mesmo sem dados completos do cliente.
 */

// Habilita o log de erros
ini_set('display_errors', 0);
ini_set('log_errors', 1);
error_reporting(E_ALL);

header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');

// Configurações

/**
 * Função auxiliar para processar valor dinâmico
 * @param mixed $input_valor Valor recebido (GET/POST)
 * @param int $default_valor Valor padrão em centavos
 * @return int Valor em centavos
 */
function processarValor($input_valor, $default_valor = 13845) {
    if (empty($input_valor)) {
        return $default_valor;
    }
    
    $valor = floatval($input_valor);
    
    // Se o valor é menor que 100 e maior que 0, assume que está em reais
    if ($valor < 100 && $valor > 0) {
        $valor = $valor * 100; // Converte para centavos
    }
    
    return intval($valor);
}

/**
 * Valida CPF
 * @param string $cpf CPF para validar (apenas números)
 * @return bool True se válido, False se inválido
 */
function validarCPF($cpf) {
    // Remove caracteres não numéricos
    $cpf = preg_replace('/[^0-9]/', '', $cpf);
    
    // Verifica se tem 11 dígitos
    if (strlen($cpf) != 11) {
        return false;
    }
    
    // Verifica se todos os dígitos são iguais
    if (preg_match('/(\d)\1{10}/', $cpf)) {
        return false;
    }
    
    // Valida primeiro dígito verificador
    for ($t = 9; $t < 11; $t++) {
        for ($d = 0, $c = 0; $c < $t; $c++) {
            $d += $cpf[$c] * (($t + 1) - $c);
        }
        $d = ((10 * $d) % 11) % 10;
        if ($cpf[$c] != $d) {
            return false;
        }
    }
    
    return true;
}
$secretKey = "sk_Fesv-MjELX0JsSV7oll1i7yYQaC9tZDFsDmcYFSXxiL6QIey";
$apiUrl = "https://api.quantumpayments.com.br/v1/transactions";

// Array para armazenar logs
$logs = [];
$logs[] = "Iniciando processamento de pagamento PIX";

// Função para gerar CPF válido
function gerarCPF() {
    $cpf = '';
    for ($i = 0; $i < 9; $i++) {
        $cpf .= rand(0, 9);
    }

    $soma = 0;
    for ($i = 0; $i < 9; $i++) {
        $soma += intval($cpf[$i]) * (10 - $i);
    }
    $resto = $soma % 11;
    $digito1 = ($resto < 2) ? 0 : 11 - $resto;
    $cpf .= $digito1;

    $soma = 0;
    for ($i = 0; $i < 10; $i++) {
        $soma += intval($cpf[$i]) * (11 - $i);
    }
    $resto = $soma % 11;
    $digito2 = ($resto < 2) ? 0 : 11 - $resto;
    $cpf .= $digito2;

    $invalidos = [
        '00000000000', '11111111111', '22222222222', '33333333333', 
        '44444444444', '55555555555', '66666666666', '77777777777', 
        '88888888888', '99999999999'
    ];

    if (in_array($cpf, $invalidos)) {
        return gerarCPF();
    }

    return $cpf;
}

/**
 * Função para gerar um email fictício baseado no nome
 */
function gerarEmail($nome) {
    $nome = strtolower(trim($nome));
    $nome = preg_replace('/[^a-z0-9]/', '', iconv('UTF-8', 'ASCII//TRANSLIT', $nome));
    $dominios = ['gmail.com', 'hotmail.com', 'outlook.com', 'yahoo.com.br', 'uol.com.br'];
    $dominio = $dominios[array_rand($dominios)];
    
    return $nome . rand(1, 999) . '@' . $dominio;
}

try {
    // Conecta ao SQLite (arquivo de banco de dados)
    $dbPath = __DIR__ . '/database.sqlite'; // Caminho para o arquivo SQLite
    $db = new PDO("sqlite:$dbPath");
    $db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $logs[] = "Conexão com banco de dados SQLite estabelecida: $dbPath";

    // Verifica se a tabela 'pedidos' existe e cria se necessário
    $db->exec("CREATE TABLE IF NOT EXISTS pedidos (
        transaction_id TEXT PRIMARY KEY,
        status TEXT NOT NULL,
        valor INTEGER NOT NULL,
        nome TEXT,
        email TEXT,
        cpf TEXT,
        pix_code TEXT,
        utm_params TEXT,
        created_at TEXT,
        updated_at TEXT
    )");
    $logs[] = "Tabela 'pedidos' verificada/criada com sucesso";
    
    // Adiciona a coluna pix_code se não existir (para bancos antigos)
    try {
        $db->exec("ALTER TABLE pedidos ADD COLUMN pix_code TEXT");
        $logs[] = "Coluna 'pix_code' adicionada à tabela 'pedidos'";
    } catch (PDOException $e) {
        // Coluna já existe, ignora o erro
        $logs[] = "Coluna 'pix_code' já existe na tabela 'pedidos'";
    }

    // Valor dinâmico em centavos
    $valor_default = 2167; // Valor padrão em centavos (R$ 59,40)
    
    // Processa valor dinâmico
    $input_valor = $_GET['valor'] ?? $_POST['valor'] ?? null;
    $valor = processarValor($input_valor, $valor_default);
    $valor_centavos = $valor;
    
    // Log do processamento
    if (!empty($input_valor)) {
        $metodo = isset($_GET['valor']) ? 'GET' : 'POST';
        $logs[] = "Valor recebido via $metodo: $input_valor";
    } else {
        $logs[] = "Usando valor padrão: $valor_default centavos";
    }
    
    $logs[] = "Valor final processado: $valor centavos (R$ " . number_format($valor/100, 2, ',', '.') . ")";

    // Validação do valor
    if (!$valor || $valor <= 0) {
        $logs[] = "ERRO: Valor inválido: $valor";
        throw new Exception('Valor inválido. O valor deve ser maior que zero.');
    }
    
    if ($valor < 100) { // Menos de R$ 1,00
        $logs[] = "AVISO: Valor muito baixo: $valor centavos";
    }
    
    if ($valor > 1000000) { // Mais de R$ 10.000,00
        $logs[] = "AVISO: Valor muito alto: $valor centavos";
    }

    // Dados falsos para fallback
    $nomes_masculinos = [
        'João', 'Pedro', 'Lucas', 'Miguel', 'Arthur', 'Gabriel', 'Bernardo', 'Rafael',
        'Gustavo', 'Felipe', 'Daniel', 'Matheus', 'Bruno', 'Thiago', 'Carlos'
    ];

    $nomes_femininos = [
        'Maria', 'Ana', 'Julia', 'Sofia', 'Isabella', 'Helena', 'Valentina', 'Laura',
        'Alice', 'Manuela', 'Beatriz', 'Clara', 'Luiza', 'Mariana', 'Sophia'
    ];

    $sobrenomes = [
        'Silva', 'Santos', 'Oliveira', 'Souza', 'Rodrigues', 'Ferreira', 'Alves', 
        'Pereira', 'Lima', 'Gomes', 'Costa', 'Ribeiro', 'Martins', 'Carvalho', 
        'Almeida', 'Lopes', 'Soares', 'Fernandes', 'Vieira', 'Barbosa'
    ];

    // Parâmetros UTM
    $utmParams = [
        'utm_source' => $_POST['utm_source'] ?? null,
        'utm_medium' => $_POST['utm_medium'] ?? null,
        'utm_campaign' => $_POST['utm_campaign'] ?? null,
        'utm_content' => $_POST['utm_content'] ?? null,
        'utm_term' => $_POST['utm_term'] ?? null,
        'xcod' => $_POST['xcod'] ?? null,
        'sck' => $_POST['sck'] ?? null
    ];

    $utmParams = array_filter($utmParams, function($value) {
        return $value !== null && $value !== '';
    });
    $logs[] = "Parâmetros UTM recebidos: " . json_encode($utmParams);

    // Recebe dados do cliente (GET ou POST)
    $nome_cliente_input = $_GET['nome'] ?? $_POST['nome'] ?? $_POST['name'] ?? null;
    $cpf_input = $_GET['cpf'] ?? $_POST['cpf'] ?? $_POST['document'] ?? null;
    $telefone_input = $_GET['telefone'] ?? $_POST['telefone'] ?? $_POST['telephone'] ?? $_POST['phone'] ?? null;
    $email_input = $_GET['email'] ?? $_POST['email'] ?? null;
    
    // Processa e valida dados recebidos
    $usar_dados_reais = false;
    
    // Nome
    if (!empty($nome_cliente_input) && strlen(trim($nome_cliente_input)) >= 3) {
        $nome_cliente = trim($nome_cliente_input);
        $usar_dados_reais = true;
        $logs[] = "Nome recebido: $nome_cliente";
    } else {
        // Gera nome falso como fallback
        $genero = rand(0, 1);
        $nome = $genero ? 
            $nomes_masculinos[array_rand($nomes_masculinos)] : 
            $nomes_femininos[array_rand($nomes_femininos)];
        $sobrenome1 = $sobrenomes[array_rand($sobrenomes)];
        $sobrenome2 = $sobrenomes[array_rand($sobrenomes)];
        $nome_cliente = "$nome $sobrenome1 $sobrenome2";
        $logs[] = "Nome gerado (fallback): $nome_cliente";
    }
    
    // CPF
    if (!empty($cpf_input)) {
        // Remove formatação do CPF
        $cpf = preg_replace('/[^0-9]/', '', $cpf_input);
        if (validarCPF($cpf)) {
            $usar_dados_reais = true;
            $logs[] = "CPF recebido e validado: $cpf";
        } else {
            $cpf = gerarCPF();
            $logs[] = "CPF inválido recebido ($cpf_input), gerando novo (fallback): $cpf";
        }
    } else {
        $cpf = gerarCPF();
        $logs[] = "CPF gerado (fallback): $cpf";
    }
    
    // Telefone
    if (!empty($telefone_input)) {
        // Remove formatação do telefone
        $telefone = preg_replace('/[^0-9]/', '', $telefone_input);
        if (strlen($telefone) >= 10 && strlen($telefone) <= 11) {
            // Formata telefone
            if (strlen($telefone) === 11) {
                $telefone_formatado = '(' . substr($telefone, 0, 2) . ') ' . substr($telefone, 2, 5) . '-' . substr($telefone, 7);
            } else {
                $telefone_formatado = '(' . substr($telefone, 0, 2) . ') ' . substr($telefone, 2, 4) . '-' . substr($telefone, 6);
            }
            $usar_dados_reais = true;
            $logs[] = "Telefone recebido: $telefone_formatado";
        } else {
            $telefone_formatado = '(11) 99999-9999';
            $logs[] = "Telefone inválido recebido, usando padrão (fallback): $telefone_formatado";
        }
    } else {
        $telefone_formatado = '(11) 99999-9999';
        $logs[] = "Telefone gerado (fallback): $telefone_formatado";
    }
    
    // Email
    if (!empty($email_input) && filter_var($email_input, FILTER_VALIDATE_EMAIL)) {
        $email = trim($email_input);
        $usar_dados_reais = true;
        $logs[] = "Email recebido: $email";
    } else {
        // Gera email baseado no nome
        $email = gerarEmail($nome_cliente);
        $logs[] = "Email gerado (fallback): $email";
    }
    
    // Placa (sempre gerada)
    $placa = chr(rand(65, 90)) . chr(rand(65, 90)) . chr(rand(65, 90)) . rand(0, 9) . rand(0, 9) . rand(0, 9) . rand(0, 9);
    
    // Log final
    if ($usar_dados_reais) {
        $logs[] = "✓ Usando dados REAIS do cliente";
    } else {
        $logs[] = "⚠ Usando dados FALSOS (fallback) - nenhum dado válido foi recebido";
    }
    $logs[] = "Dados finais: nome=$nome_cliente, cpf=$cpf, telefone=$telefone_formatado, email=$email, placa=$placa";
    
    // Formatar valor para exibição
    $valorFormatado = 'R$ ' . number_format($valor_centavos/100, 2, ',', '.');
    $logs[] = "Valor formatado: $valorFormatado";

    // Preparar dados para a API
    $data = [
        'amount' => $valor_centavos, // Valor em unidades inteiras
        'paymentMethod' => 'pix', // Definindo o método de pagamento como PIX
        'pix' => [
            'expiresInDays' => 1 // Expira em 1 dia
        ],
        'customer' => [
            'name' => $nome_cliente,
            'email' => $email,
            'phone' => $telefone_formatado, // Telefone dinâmico
            'document' => [
                'type' => 'cpf',
                'number' => $cpf
            ],
            'externalRef' => 'md-' . $placa . '-' . time() // Referência externa
        ],
        'items' => [
            [
                'title' => 'Curso helton vieira modulo' . $placa,
                'unitPrice' => $valor_centavos, // Valor em unidades inteiras
                'quantity' => 1,
                'tangible' => false,
                'externalRef' => 'IPVA-' . $placa
            ]
        ],
        'metadata' => json_encode($utmParams),
        'ip' => $_SERVER['REMOTE_ADDR'] ?? '127.0.0.1'
    ];
    
    $logs[] = "Payload para API: " . json_encode($data);
    
    // Fazer requisição para a API
    $authorization = 'Basic ' . base64_encode($secretKey . ':x');
    $logs[] = "Authorization: Basic ***********";
    
    // Fazer a requisição real para a API
    $ch = curl_init($apiUrl);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        'Authorization: ' . $authorization,
        'Content-Type: application/json',
        'Accept: application/json'
    ]);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    
    $response = curl_exec($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $curlError = curl_error($ch);
    curl_close($ch);
    
    $logs[] = "Resposta da API - HTTP Code: $httpCode";
    if (!empty($curlError)) {
        $logs[] = "Erro cURL: $curlError";
        throw new Exception("Erro cURL: $curlError");
    }
    
    if ($response) {
        $logs[] = "Resposta bruta: " . $response;
    } else {
        $logs[] = "Sem resposta da API";
        throw new Exception("Sem resposta da API");
    }
    
    if ($httpCode === 200 || $httpCode === 201) {
        $responseData = json_decode($response, true);
        $logs[] = "Resposta decodificada: " . json_encode($responseData);
        
        if (!isset($responseData['id'])) {
            throw new Exception("ID não encontrado na resposta da API");
        }
        
        // Extrair os dados do PIX da resposta
        // Verificamos todos os possíveis campos onde pode estar o código PIX
        $pixCopiaECola = '';
        if (isset($responseData['pix']['qrcode'])) {
            $pixCopiaECola = $responseData['pix']['qrcode'];
            $logs[] = "Código PIX encontrado em responseData['pix']['qrcode']";
        } elseif (isset($responseData['pix']['qrCode'])) {
            $pixCopiaECola = $responseData['pix']['qrCode'];
            $logs[] = "Código PIX encontrado em responseData['pix']['qrCode']";
        } elseif (isset($responseData['pix']['code'])) {
            $pixCopiaECola = $responseData['pix']['code'];
            $logs[] = "Código PIX encontrado em responseData['pix']['code']";
        } elseif (isset($responseData['pix']['text'])) {
            $pixCopiaECola = $responseData['pix']['text'];
            $logs[] = "Código PIX encontrado em responseData['pix']['text']";
        } elseif (isset($responseData['qrcode'])) {
            $pixCopiaECola = $responseData['qrcode'];
            $logs[] = "Código PIX encontrado em responseData['qrcode']";
        }
        
        // Fazer o mesmo para a URL do QR Code
        $qrCodeUrl = '';
        if (isset($responseData['pix']['receiptUrl'])) {
            $qrCodeUrl = $responseData['pix']['receiptUrl'];
            $logs[] = "URL do QR Code encontrado em responseData['pix']['receiptUrl']";
        } elseif (isset($responseData['pix']['qrcodeUrl'])) {
            $qrCodeUrl = $responseData['pix']['qrcodeUrl'];
            $logs[] = "URL do QR Code encontrado em responseData['pix']['qrcodeUrl']";
        } elseif (isset($responseData['pix']['imageUrl'])) {
            $qrCodeUrl = $responseData['pix']['imageUrl'];
            $logs[] = "URL do QR Code encontrado em responseData['pix']['imageUrl']";
        } elseif (isset($responseData['qrcodeUrl'])) {
            $qrCodeUrl = $responseData['qrcodeUrl'];
            $logs[] = "URL do QR Code encontrado em responseData['qrcodeUrl']";
        }
        
        $txid = isset($responseData['pix']['end2EndId']) ? $responseData['pix']['end2EndId'] : '';
        if (empty($txid) && isset($responseData['pix']['txid'])) {
            $txid = $responseData['pix']['txid'];
        }
        
        $logs[] = "Dados PIX extraídos - qrCode: " . (empty($pixCopiaECola) ? 'vazio' : 'preenchido');
        $logs[] = "Dados PIX extraídos - qrCodeUrl: " . (empty($qrCodeUrl) ? 'vazio' : 'preenchido');
        $logs[] = "Dados PIX extraídos - txid: " . (empty($txid) ? 'vazio' : $txid);
        
        // Se não conseguir obter os dados da API, informar erro
      
        
        // Gerar QR Code usando o QRServer
        $qrCodeImageUrl = "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" . urlencode($pixCopiaECola);
        $logs[] = "URL do QR Code gerado: $qrCodeImageUrl";
        
        // Verificar se já existe um registro com este transaction_id
        $checkStmt = $db->prepare("SELECT COUNT(*) FROM pedidos WHERE transaction_id = :transaction_id");
        $checkStmt->execute(['transaction_id' => $responseData['id']]);
        $exists = (int)$checkStmt->fetchColumn() > 0;
        
        if ($exists) {
            $logs[] = "Pedido já existe no banco de dados. Atualizando informações.";
            $stmt = $db->prepare("UPDATE pedidos SET 
                status = :status,
                pix_code = :pix_code,
                updated_at = :updated_at 
                WHERE transaction_id = :transaction_id");
            $stmt->execute([
                'status' => 'pending',
                'pix_code' => $pixCopiaECola,
                'transaction_id' => $responseData['id'],
                'updated_at' => date('c')
            ]);
        } else {
            // Salva os dados no SQLite - Garantindo que dados sejam inseridos corretamente
            $logs[] = "Inserindo novo registro no banco de dados: " . $responseData['id'];
            try {
                $stmt = $db->prepare("INSERT INTO pedidos (transaction_id, status, valor, nome, email, cpf, pix_code, utm_params, created_at) 
                    VALUES (:transaction_id, :status, :valor, :nome, :email, :cpf, :pix_code, :utm_params, :created_at)");
                $result = $stmt->execute([
                    'transaction_id' => $responseData['id'],
                    'status' => 'pending',
                    'valor' => $valor_centavos,
                    'nome' => $nome_cliente,
                    'email' => $email,
                    'cpf' => $cpf,
                    'pix_code' => $pixCopiaECola,
                    'utm_params' => json_encode($utmParams),
                    'created_at' => date('c')
                ]);
                
                if ($result) {
                    $logs[] = "Dados salvos com sucesso no banco de dados SQLite";
                } else {
                    $logs[] = "ERRO: Falha ao inserir dados no banco de dados";
                }
            } catch (PDOException $e) {
                $logs[] = "ERRO de banco de dados: " . $e->getMessage();
                // Não interrompe o fluxo, apenas registra o erro
            }
        }
        
        // Garantir que a sessão está ativa
        if (session_status() === PHP_SESSION_NONE) {
            session_start();
        }
        $_SESSION['payment_id'] = $responseData['id'];
        $logs[] = "ID do pagamento salvo na sessão: " . $responseData['id'];
        
        // Enviar para utmify-pendente.php
        $utmifyData = [
            'orderId' => $responseData['id'],
            'platform' => 'BlackCat',
            'paymentMethod' => 'pix',
            'status' => 'waiting_payment',
            'createdAt' => date('Y-m-d H:i:s'),
            'approvedDate' => null,
            'refundedAt' => null,
            'customer' => [
                'name' => $nome_cliente,
                'email' => $email,
                'phone' => null,
                'document' => $cpf,
                'country' => 'BR',
                'ip' => $_SERVER['REMOTE_ADDR'] ?? null
            ],
            'products' => [
                [
                    'id' => 'PROD_' . rand(1000, 9999),
                    'name' => 'Curso helton vieira',
                    'planId' => null,
                    'planName' => null,
                    'quantity' => 1,
                    'priceInCents' => $valor_centavos
                ]
            ],
            'trackingParameters' => $utmParams,
            'commission' => [
                'totalPriceInCents' => $valor_centavos,
                'gatewayFeeInCents' => isset($responseData['fee']['amount']) ? $responseData['fee']['amount'] : 0,
                'userCommissionInCents' => $valor_centavos
            ],
            'isTest' => false
        ];
        
        $logs[] = "Dados preparados para utmify-pendente.php";
        
        $serverUrl = (isset($_SERVER['HTTPS']) && $_SERVER['HTTPS'] === 'on' ? "https" : "http") . "://$_SERVER[HTTP_HOST]";
        $utmifyUrl = $serverUrl . "/utmify-pendente.php";
        $logs[] = "URL do utmify-pendente.php: " . $utmifyUrl;
        
        $ch = curl_init($utmifyUrl);
        curl_setopt_array($ch, [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_POST => true,
            CURLOPT_POSTFIELDS => json_encode($utmifyData),
            CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_SSL_VERIFYHOST => false
        ]);
        
        $utmifyResponse = curl_exec($ch);
        $utmifyHttpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $utmifyError = curl_error($ch);
        curl_close($ch);
        
        $logs[] = "Resposta do utmify-pendente.php - HTTP Code: $utmifyHttpCode";
        if (!empty($utmifyError)) {
            $logs[] = "Erro ao enviar para utmify-pendente.php: $utmifyError";
        }
        
        if ($utmifyHttpCode !== 200) {
            $logs[] = "Alerta: Resposta não-200 do utmify-pendente.php: " . $utmifyResponse;
        } else {
            $logs[] = "Dados enviados com sucesso para utmify-pendente.php";
        }
        
        // Retornar dados para o frontend
        $responseToFrontend = [
            'success' => true,
            'qrCodeUrl' => $qrCodeImageUrl, // URL do QR Code gerado
            'pixCode' => $pixCopiaECola,
            'valor' => $valorFormatado,
            'nome' => $nome_cliente,
            'cpf' => $cpf,
            'placa' => $placa,
            'expiraEm' => '1 dia',
            'txid' => $txid,
            'token' => $responseData['id'], // Adicionado para garantir que o frontend tenha o ID
            'logs' => $logs
        ];
        
        $logs[] = "Enviando resposta de sucesso para o frontend";
        echo json_encode($responseToFrontend);
    } else {
        // Tratar erro
        $errorMessage = 'Erro ao processar pagamento';
        $errorDetails = '';
        
        if ($response) {
            $responseData = json_decode($response, true);
            $logs[] = "Resposta de erro decodificada: " . json_encode($responseData);
            
            if (isset($responseData['message'])) {
                $errorMessage = $responseData['message'];
                $logs[] = "Mensagem de erro da API: $errorMessage";
            }
            
            // Capturar detalhes do erro
            if (isset($responseData['details'])) {
                $errorDetails = is_array($responseData['details']) ? 
                    json_encode($responseData['details']) : 
                    $responseData['details'];
                $logs[] = "Detalhes do erro: $errorDetails";
            }
        }
        
        throw new Exception($errorMessage . ($errorDetails ? ": " . $errorDetails : ""));
    }
} catch (Exception $e) {
    $logs[] = "❌ Erro: " . $e->getMessage();
    $logs[] = "🔍 Stack trace: " . $e->getTraceAsString();
    
    echo json_encode([
        'success' => false,
        'message' => 'Erro ao gerar o PIX: ' . $e->getMessage(),
        'logs' => $logs
    ]);
}
?>